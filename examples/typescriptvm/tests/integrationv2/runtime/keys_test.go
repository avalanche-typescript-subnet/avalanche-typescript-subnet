package runtime_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteKeys(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateprovider := &DummyStateProvider{
		State: map[runtime.KeyPostfix][]byte{},
	}

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{4},
		FunctionName:  "writeManySlots",
	}
	var callback runtime.CallbackFunc = func(src []byte) ([]byte, error) {
		dst := make([]byte, len(src)*2)
		for i := 0; i < len(src)*2; i++ {
			if i < len(src) {
				dst[i] = src[i]
			} else {
				dst[i] = byte(100 + i)
			}
		}
		return dst, nil
	}

	res, err := exec.Execute(params, callback)
	if err != nil {
		t.Fatal(err)
	}
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateprovider.SetState(updatedKey, updatedVal)
	}

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//check write keys
	expectedWriteKeys := map[runtime.KeyPostfix][]byte{
		{0, 1, 0, 6}: {1, 1, 1}, //starting with slot 1, size is fixed to 6
		{0, 2, 0, 6}: {2, 2, 2},
		{0, 3, 0, 6}: {3, 3, 3},
		{0, 4, 0, 6}: {4, 4, 4},
	}

	require.Equal(t, expectedWriteKeys, res.Result.UpdatedKeys)

	assert.Equal(t, len(expectedWriteKeys), len(res.Result.UpdatedKeys), "Updated keys count mismatch")
}

func TestReadKeys(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateprovider := &DummyStateProvider{
		State: map[runtime.KeyPostfix][]byte{},
	}

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{4},
		FunctionName:  "writeManySlots",
	}

	res, err := exec.Execute(params, runtime.EmptyCallbackFunc)
	require.NoError(t, err)
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateprovider.SetState(updatedKey, updatedVal)
	}

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//read
	params.Payload = []byte{4}
	params.FunctionName = "readManySlots"
	res, err = exec.Execute(params, runtime.EmptyCallbackFunc)
	require.NoError(t, err)

	require.Equal(t, true, res.Result.Success, res.Result.Error)
	require.Equal(t, 4, len(res.Result.ReadKeys))
	expectedReadKeys := []runtime.KeyPostfix{
		{0, 1, 0, 6},
		{0, 2, 0, 6},
		{0, 3, 0, 6},
		{0, 4, 0, 6},
	}
	require.Equal(t, expectedReadKeys, res.Result.ReadKeys)
}
