package wasmsdk_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/cmd/wasmsdk/wasmsdk"
	"github.com/stretchr/testify/require"
)

func TestBaseSerializeJSON(t *testing.T) {
	base := wasmsdk.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5, 6},
		MaxFee:    123456789,
	}

	serialized, err := json.Marshal(base)
	require.NoError(t, err)

	var deserialized wasmsdk.Base
	err = json.Unmarshal(serialized, &deserialized)
	require.NoError(t, err)

	require.Equal(t, base.Timestamp, deserialized.Timestamp)
	require.Equal(t, base.ChainID, deserialized.ChainID)
	require.Equal(t, base.MaxFee, deserialized.MaxFee)
}
