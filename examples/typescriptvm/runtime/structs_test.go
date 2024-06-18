package runtime_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/stretchr/testify/require"
)

func TestContractStateMap_UnmarshalJSON(t *testing.T) {
	inputJson := `{"success":true,"result":"","readKeys":[],"updatedKeys":{"AQAG":"AQEB","AgAG":"AgIC","AwAG":"AwMD","BAAG":"BAQE"}}`

	var stdoutResult runtime.ResultJSON
	err := json.Unmarshal([]byte(inputJson), &stdoutResult)
	require.NoError(t, err)

	var key []byte
	var value []byte
	var exists bool

	key = base64ToBytes("AQAG")
	value, exists = stdoutResult.UpdatedKeys.Get(key)
	require.True(t, exists)
	require.Equal(t, base64ToBytes("AQEB"), value)

	key = base64ToBytes("AgAG")
	value, exists = stdoutResult.UpdatedKeys.Get(key)
	require.True(t, exists)
	require.Equal(t, base64ToBytes("AgIC"), value)

	key = base64ToBytes("AwAG")
	value, exists = stdoutResult.UpdatedKeys.Get(key)
	require.True(t, exists)
	require.Equal(t, base64ToBytes("AwMD"), value)

	key = base64ToBytes("BAAG")
	value, exists = stdoutResult.UpdatedKeys.Get(key)
	require.True(t, exists)
	require.Equal(t, base64ToBytes("BAQE"), value)

	require.Equal(t, 4, len(stdoutResult.UpdatedKeys.Data()))
}

func base64ToBytes(base64Str string) []byte {
	bytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		panic(err)
	}
	return bytes
}
