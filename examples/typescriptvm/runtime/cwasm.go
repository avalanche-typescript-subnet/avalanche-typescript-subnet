package runtime

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/bytecodealliance/wasmtime-go/v21"
)

const WASMTIME_VERSION = "v21"

//go:generate bash -c "test -f ./javy_provider.wasm || curl -L https://github.com/bytecodealliance/javy/releases/download/v3.0.0/javy-quickjs_provider.wasm.gz | gunzip > ./javy_provider.wasm"

//go:embed javy_provider.wasm
var javyProviderWasm []byte

var javyProviderCompiled *[]byte = nil

var compileWasmMutex = sync.Mutex{}

func getCwasmBytes() (*[]byte, error) {
	if javyProviderCompiled != nil {
		return javyProviderCompiled, nil
	}

	compileWasmMutex.Lock()
	defer compileWasmMutex.Unlock()

	if javyProviderCompiled != nil {
		return javyProviderCompiled, nil
	}

	config := wasmtime.NewConfig()
	config.SetConsumeFuel(true)

	engine := wasmtime.NewEngineWithConfig(config)

	javyProviderModule, err := wasmtime.NewModule(engine, javyProviderWasm)
	if err != nil {
		return nil, fmt.Errorf("instantiating javy provider module: %v", err)
	}

	compiledJavyProviderWasm, err := javyProviderModule.Serialize()
	if err != nil {
		return nil, fmt.Errorf("serializing javy provider module: %v", err)
	}

	javyProviderCompiled = &compiledJavyProviderWasm

	return javyProviderCompiled, nil
}

func getCwasmCachePath() (string, error) {
	return "", nil
}
