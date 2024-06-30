package wasmsdk_test

import (
	"testing"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/cmd/wasmsdk/wasmsdk"
	"github.com/stretchr/testify/require"
)

func TestDigestNil(t *testing.T) {
	timestamp := int64(1719282804 * 1000)

	base := chain.Base{
		Timestamp: timestamp,
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}
	fullTx := chain.NewTx(&base, nil)
	originalDigest, err := fullTx.Digest()
	require.NoError(t, err)

	compactTx := wasmsdk.Transaction{
		Base: &wasmsdk.Base{
			Timestamp: base.Timestamp,
			ChainID:   base.ChainID,
			MaxFee:    base.MaxFee,
		},
	}

	wasmDigest, err := compactTx.Digest()
	require.NoError(t, err)

	require.Equal(t, originalDigest, wasmDigest)
}

func TestDigestTransfer(t *testing.T) {
	timestamp := int64(1719282804 * 1000)

	base := chain.Base{
		Timestamp: timestamp,
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

	fullTx := chain.NewTx(&base, actions)
	originalDigest, err := fullTx.Digest()
	require.NoError(t, err)

	compactTx := wasmsdk.Transaction{
		Base: &wasmsdk.Base{
			Timestamp: base.Timestamp,
			ChainID:   base.ChainID,
			MaxFee:    base.MaxFee,
		},
		Actions: []wasmsdk.CompactAction{
			actions[0].(wasmsdk.CompactAction),
			actions[1].(wasmsdk.CompactAction),
		},
	}

	wasmDigest, err := compactTx.Digest()
	require.NoError(t, err)

	require.Equal(t, originalDigest, wasmDigest)
}
