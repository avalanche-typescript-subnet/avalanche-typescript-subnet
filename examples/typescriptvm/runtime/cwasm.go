package runtime

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/bytecodealliance/wasmtime-go/v21"
)

const WASMTIME_VERSION = "v21"

//go:embed javy_provider_1.4.0.wasm
var javyProviderWasm []byte

var javyProviderCompiled *[]byte = nil

var compileWasmMutex = sync.Mutex{}

func getCwasmBytes() (*[]byte, error) {
	if javyProviderCompiled != nil {
		return javyProviderCompiled, nil
	}

	cachedFilePath, err := getCwasmCachePath()
	if err != nil {
		return nil, fmt.Errorf("getting cwasm cache path: %v", err)
	}

	compileWasmMutex.Lock()
	defer compileWasmMutex.Unlock()

	_, err = os.Stat(cachedFilePath)

	if err != nil && os.IsNotExist(err) {
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

		if err := os.WriteFile(cachedFilePath, compiledJavyProviderWasm, 0644); err != nil {
			return nil, fmt.Errorf("writing cwasm to cache: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("checking if cwasm cache exists: %v", err)
	}

	_javyProviderCompiled, err := os.ReadFile(cachedFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading cwasm from cache: %v", err)
	}

	javyProviderCompiled = &_javyProviderCompiled

	return javyProviderCompiled, nil
}

func getCwasmCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("getting user cache dir: %v", err)
	}

	result := filepath.Join(cacheDir, fmt.Sprintf("javy_provider_%s.cwasm", WASMTIME_VERSION))
	return result, nil
}
