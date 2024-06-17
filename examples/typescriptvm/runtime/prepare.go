package runtime

import (
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/bytecodealliance/wasmtime-go/v21"
)

// speeds up code execution 3 times in exchange for higher memory usage and and non-deterministic gas counting.
const EXECUTE_CACHE_ENABLED = false

//go:generate bash -c "cd js_sdk && [ ! -d node_modules ] && npm ci || echo 'node_modules already present'"

func (exec *JavyExec) createStore(wasmBytes *[]byte) (*wasmtime.Store, *wasmtime.Func, error) {

	var wasmHash uint64
	if EXECUTE_CACHE_ENABLED {
		wasmHash = hashBytes(*wasmBytes)

		if cached, found := exec.storesCache[wasmHash]; found {
			return cached.store, cached.mainFunc, nil
		}
	}

	config := wasmtime.NewConfig()
	config.SetConsumeFuel(true)

	engine := wasmtime.NewEngineWithConfig(config)

	userCodeModule, err := wasmtime.NewModule(engine, *wasmBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating user code module: %v", err)
	}

	compiledLib, err := getCwasmBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("getting javy provider compiled wasm: %v", err)
	}

	libraryModule, err := wasmtime.NewModuleDeserialize(engine, *compiledLib)
	if err != nil {
		cwasmCachePath, _ := getCwasmCachePath()
		log.Printf("Library size: %d", len(*compiledLib))
		return nil, nil, fmt.Errorf("instantiating javy library module (consider cleaning up %s): %v", cwasmCachePath, err)
	}

	store := wasmtime.NewStore(engine)

	linker := wasmtime.NewLinker(engine)
	linker.AllowShadowing(true)
	linker.DefineWasi()

	var memory *wasmtime.Memory
	var realloc_fn *wasmtime.Func

	if callback != nil {
		err = linker.DefineFunc(store, "env", "__callback", func(argPorinter int32, argLen int32) int64 {
			pack := func(a int32, b int32) int64 {
				return int64(a)<<32 | int64(b)
			}

			wasmMem := memory.UnsafeData(store)
			arg := wasmMem[int(argPorinter) : int(argPorinter)+int(argLen)]

			res, err := callback(arg)
			if err != nil {
				fmt.Println("callback call error:", err)
				return 0
			}

			if res == nil {
				res = []byte{}
			}

			_dstPointer, err := realloc_fn.Call(store, int32(0), int32(0), int32(1), int32(len(res)))
			if err != nil {
				fmt.Println("realloc_fn call error:", err)
				return 0
			}

			dstPointer, ok := _dstPointer.(int32)
			if !ok {
				fmt.Println("dstPointer type error:", err)
				return 0
			}

			size := copy(wasmMem[dstPointer:], res)
			if size != len(res) {
				fmt.Println("copy error:", err)
				return 0
			}

			return pack(dstPointer, int32(len(res)))
		})
	} else {
		err = linker.DefineFunc(store, "env", "__callback", _emptyCallbackFunc)
	}

	if err != nil {
		fmt.Println("defining callback func wrapper:", err)
	}

	libraryInstance, err := linker.Instantiate(store, libraryModule)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating javy library instance: %v", err)
	}

	linker.DefineInstance(store, "javy_quickjs_provider_v2", libraryInstance)

	// linker.AllowShadowing(true)
	userCodeInstance, err := linker.Instantiate(store, userCodeModule)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating user code instance: %v", err)
	}

	extern := libraryInstance.GetExport(store, "memory")
	if extern == nil {
		return nil, nil, fmt.Errorf("no wasm memoryfound")
	}
	memory = extern.Memory()

	realloc_fn = libraryInstance.GetFunc(store, "canonical_abi_realloc")
	if realloc_fn == nil {
		return nil, nil, fmt.Errorf("no canonical_abi_realloc function found")
	}

	userCodeMain := userCodeInstance.GetFunc(store, "_start")

	if EXECUTE_CACHE_ENABLED {
		// Cache the store and function
		exec.storesCache[wasmHash] = &struct {
			store    *wasmtime.Store
			mainFunc *wasmtime.Func
		}{store, userCodeMain}
	}

	return store, userCodeMain, nil
}

func hashBytes(bytes []byte) uint64 {
	hash := sha256.Sum256(bytes)
	return binary.LittleEndian.Uint64(hash[:8])
}
