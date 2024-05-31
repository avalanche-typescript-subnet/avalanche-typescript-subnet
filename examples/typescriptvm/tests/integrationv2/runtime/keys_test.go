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
		State: map[[4]byte][]byte{},
	}

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{3, 4}, //ACTION_WRITE_MANY_SLOTS, value=4
	}

	res, err := exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateprovider.SetState(updatedKey, updatedVal)
	}

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//check write keys
	expectedWriteKeys := map[[4]byte][]byte{
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
		State: map[[4]byte][]byte{},
	}

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{3, 4}, //ACTION_WRITE_MANY_SLOTS, value=4
	}

	res, err := exec.Execute(params)
	require.NoError(t, err)
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateprovider.SetState(updatedKey, updatedVal)
	}

	//check read keys
	require.Equal(t, 0, len(res.Result.ReadKeys))

	//read
	params.Payload = []byte{4, 4} //ACTION_READ_MANY_SLOTS, value=4
	res, err = exec.Execute(params)
	require.NoError(t, err)

	require.Equal(t, 4, len(res.Result.ReadKeys))
	expectedReadKeys := [][4]byte{
		{0, 1, 0, 6},
		{0, 2, 0, 6},
		{0, 3, 0, 6},
		{0, 4, 0, 6},
	}
	require.Equal(t, expectedReadKeys, res.Result.ReadKeys)
}
