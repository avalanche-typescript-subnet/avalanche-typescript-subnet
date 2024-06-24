package wasmsdk_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/cmd/wasmsdk/wasmsdk"
	"github.com/stretchr/testify/require"
)

func TestDigestNil(t *testing.T) {
	base := chain.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}
	tx := chain.NewTx(&base, nil)
	originalDigest, err := tx.Digest()
	require.NoError(t, err)

	wasmDigest, err := wasmsdk.Digest(&wasmsdk.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	})
	require.NoError(t, err)

	require.Equal(t, originalDigest, wasmDigest)
}
