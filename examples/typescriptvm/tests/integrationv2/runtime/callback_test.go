package runtime_test

import (
	"fmt"
	"testing"
	"time"

	_ "embed"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/stretchr/testify/assert"
)

//go:generate ../../../runtime/js_wcb_sdk/build.sh compile assets/callback_test.ts
//go:generate ../../../runtime/js_wcb_sdk/build.sh emit-provider ../../../runtime/javy_provider.wasm

//go:embed assets/callback_test.wasm
var callbackTestWasmBytes []byte

func TestCallback(t *testing.T) {

	exec := runtime.NewJavyExec()

	stateprovider := runtime.NewDummyStateProvider()

	lengthSrcArgs := 1024
	lengthRes := 1024 // notice, disable debug output to console in exec.go line 171
	repeatNum := 3

	payload := make([]byte, lengthSrcArgs)
	for i := 0; i < lengthSrcArgs; i++ {
		payload[i] = byte(i)
	}

	//payload[0] = byte(0) //return an array with length = 1
	payload[0] = byte(1) //return an array with length = lengthRes
	var dst []byte
	//define callback function
	var callback runtime.CallbackFunc = func(src []byte) ([]byte, error) {
		dst = make([]byte, lengthRes)
		for i := 0; i < lengthRes; i++ {
			if i < len(src) {
				dst[i] = src[i]
			} else {
				dst[i] = byte(100 + i)
			}
		}
		return dst, nil
	}
	exec.SetCallbackFunc(callback)

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000 * 100000, //a large amount of fuel
		MaxTime:       time.Second * 100,         //100 seconds
		MaxMemory:     -1,                        // no limit
		Bytecode:      &callbackTestWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       payload,
		FunctionName:  "test_callback",
		Callback:      callback,
	}

	for n := 0; n < repeatNum; n++ {

		startTime := time.Now()
		res, err := exec.Execute(params)
		fmt.Println("Time to execute: ", time.Since(startTime))
		assert.NoError(t, err)
		assert.Equal(t, true, res.Result.Success)

		if payload[0] == 0 {
			assert.Equal(t, 1, len(res.Result.Result))
		} else {
			assert.Equal(t, lengthRes, len(res.Result.Result))
			for i := 0; i < lengthRes; i++ {
				assert.Equal(t, dst[i]+1, res.Result.Result[i])
			}
		}
	}

}
