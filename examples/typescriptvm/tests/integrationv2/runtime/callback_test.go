package runtime_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	_ "embed"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/stretchr/testify/assert"
)

//create a docker image with javy first (read ../../../runtime/js_wcb_sdk/readme.md for more details)
//go:generate ../../../runtime/js_wcb_sdk/build.sh compile assets/callback_test.ts
//go:generate ../../../runtime/js_wcb_sdk/build.sh emit-provider ../../../runtime/javy_provider.wasm

//go:embed assets/callback_test.wasm
var callbackTestWasmBytes []byte

func TestCallback(t *testing.T) {

	exec := runtime.NewJavyExec()
	stateprovider := runtime.NewDummyStateProvider()

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &callbackTestWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{1, 2, 3, 4, 5, 6},
		FunctionName:  "test_callback",
	}

	var dst []byte
	//define callback function
	var callback runtime.CallbackFunc = func(src []byte) ([]byte, error) {
		dst = make([]byte, len(src)*2)
		for i := 0; i < len(src)*2; i++ {
			if i < len(src) {
				dst[i] = src[i]
			} else {
				dst[i] = byte(100 + i)
			}
		}
		return dst, nil
	}

	// registering the callback
	runtime.SetCallbackFunc(callback)
	for n := 0; n < 5; n++ {

		startTime := time.Now()
		res, err := exec.Execute(params)
		fmt.Println("Time to execute: ", time.Since(startTime))
		assert.NoError(t, err)
		assert.Equal(t, true, res.Result.Success)

		//fmt.Println("Result.Result: ", string(res.Result.Result))

		resArray := map[string]byte{}
		err = json.Unmarshal(res.Result.Result, &resArray)
		assert.NoError(t, err)
		for k, v := range resArray {
			i, _ := strconv.Atoi(k)
			assert.NoError(t, err)
			assert.Equal(t, dst[i], v)
		}

	}

}
