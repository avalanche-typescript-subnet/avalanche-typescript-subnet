package runtime_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
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
	keysNum := byte(99)
	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000 * 99,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateprovider.StateProvider,
		Payload:       []byte{keysNum},
		FunctionName:  "writeManySlots",
		Actor:         []byte{1, 2, 3, 4},
	}

	var callback runtime.CallbackFunc = func(address []byte) ([]byte, error) {
		addrHex := string(address)

		addressBytes, err := hex.DecodeString(strings.TrimPrefix(addrHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("error decoding address %s: %v", addrHex, err)
		}

		newVal, err := params.StateProvider(string(addressBytes))
		if err != nil {
			return nil, fmt.Errorf("error retrieving state for address %x: %v", addressBytes, err)
		}
		return newVal, nil
	}

	exec.SetCallbackFunc(callback)

	repeatNum := 1
	result := make([]time.Duration, repeatNum*2)

	for i := 0; i < repeatNum; i++ {
		timeStart := time.Now()
		res, err := exec.Execute(params)
		result[2*i] = time.Since(timeStart)
		require.NoError(t, err)
		stateprovider.Update(res.Result.UpdatedKeys)

		//check read keys
		if i == 0 {
			require.Equal(t, 0, len(res.Result.ReadKeys))

		} else {
			require.Equal(t, int(keysNum), len(res.Result.ReadKeys))
		}

		//read
		params.Payload = []byte{99}
		params.FunctionName = "readManySlots"
		timeStart = time.Now()
		res, err = exec.Execute(params)
		result[2*i+1] = time.Since(timeStart)
		require.NoError(t, err)

		require.Equal(t, true, res.Result.Success, res.Result.Error)
		require.Equal(t, int(keysNum), len(res.Result.ReadKeys))

		expectedReadKeys := make([][]byte, int(keysNum))

		for i := 0; i < int(keysNum); i++ {
			expectedReadKeys[i] = []byte{byte(i + 1), 0, 6}
		}

		require.Equal(t, expectedReadKeys, res.Result.ReadKeys)
	}

	for i := 0; i < repeatNum; i++ {
		fmt.Println("execution time write:", result[2*i], "execution time read:", result[2*i+1])
	}
}
