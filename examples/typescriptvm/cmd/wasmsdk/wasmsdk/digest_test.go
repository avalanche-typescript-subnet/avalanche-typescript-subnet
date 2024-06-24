package wasmsdk_test

import (
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
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

	wasmDigest, err := wasmsdk.Digest(wasmsdk.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}, nil)
	require.NoError(t, err)

	require.Equal(t, originalDigest, wasmDigest)
}

func TestDigestTransfer(t *testing.T) {
	base := chain.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}

	actions := []chain.Action{
		&actions.Transfer{
			To:    [33]byte{9, 8, 7, 6, 5},
			Value: 111999,
		},
		&actions.Transfer{
			To:    [33]byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			Value: 3777777777777777777,
		},
	}

	tx := chain.NewTx(&base, actions)
	originalDigest, err := tx.Digest()
	require.NoError(t, err)

	wasmDigest, err := wasmsdk.Digest(wasmsdk.Base{
		Timestamp: time.Now().Unix(),
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}, []wasmsdk.CompactAction{actions[0].(wasmsdk.CompactAction), actions[1].(wasmsdk.CompactAction)})
	require.NoError(t, err)

	require.Equal(t, originalDigest, wasmDigest)
}
