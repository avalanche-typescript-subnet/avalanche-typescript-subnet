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
	linker.DefineWasi()

	libraryInstance, err := linker.Instantiate(store, libraryModule)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating javy library instance: %v", err)
	}

	linker.DefineInstance(store, "javy_quickjs_provider_v1", libraryInstance)

	linker.AllowShadowing(true)
	userCodeInstance, err := linker.Instantiate(store, userCodeModule)
	if err != nil {
		return nil, nil, fmt.Errorf("instantiating user code instance: %v", err)
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
