package runtime_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"

	"github.com/stretchr/testify/require"
)

func TestWriteKeys(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateprovider := runtime.NewDummyStateProvider()

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{4},
		FunctionName:  "writeManySlots",
	}

	res, err := exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	stateprovider.Update(res.Result.UpdatedKeys)

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//check write keys
	expectedWriteKeys := map[string][]byte{
		string([]byte{1, 0, 6}): {1, 1, 1}, //starting with slot 1, size is fixed to 6
		string([]byte{2, 0, 6}): {2, 2, 2},
		string([]byte{3, 0, 6}): {3, 3, 3},
		string([]byte{4, 0, 6}): {4, 4, 4},
	}

	require.Equal(t, len(expectedWriteKeys), len(res.Result.UpdatedKeys), "Updated keys count mismatch")
	require.Equal(t, expectedWriteKeys, res.Result.UpdatedKeys)
}

func b64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func TestReadKeys(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateprovider := runtime.NewDummyStateProvider()

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{4},
		FunctionName:  "writeManySlots",
		Actor:         []byte{1, 2, 3, 4},
	}

	res, err := exec.Execute(params)
	require.NoError(t, err)
	stateprovider.Update(res.Result.UpdatedKeys)

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//read
	params.Payload = []byte{4}
	params.FunctionName = "readManySlots"
	res, err = exec.Execute(params)
	require.NoError(t, err)

	require.Equal(t, true, res.Result.Success, res.Result.Error)
	require.Equal(t, 4, len(res.Result.ReadKeys))
	expectedReadKeys := [][]byte{
		{1, 0, 6},
		{2, 0, 6},
		{3, 0, 6},
		{4, 0, 6},
	}

	require.Equal(t, expectedReadKeys, res.Result.ReadKeys)
}
