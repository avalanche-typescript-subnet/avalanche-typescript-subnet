package wasmsdk_test

import (
	"encoding/hex"
	"testing"

	golang_ed25519 "crypto/ed25519"

	ava_ed25519 "github.com/ava-labs/hypersdk/crypto/ed25519"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/auth"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/cmd/wasmsdk/wasmsdk"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/stretchr/testify/require"

	_ "github.com/ava-labs/hypersdk/examples/typescriptvm/registry"
)

func TestManualSign(t *testing.T) {
	timestamp := int64(1719282804 * 1000)

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

	base := chain.Base{
		Timestamp: timestamp,
		ChainID:   [32]byte{0, 1, 2, 3, 4, 5},
		MaxFee:    123456789,
	}
	fullTx := chain.NewTx(&base, actions)

	privKey, err := hex.DecodeString("323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7")
	require.NoError(t, err)

	authFactory := auth.NewED25519Factory(
		ava_ed25519.PrivateKey(privKey),
	)

	fullTx, err = fullTx.Sign(authFactory, consts.ActionRegistry, consts.AuthRegistry)
	require.NoError(t, err)

	originalSig := fullTx.Bytes()
	originalSigHex := hex.EncodeToString(originalSig)

	//our rebuild from here
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

	const ED25519ID uint8 = 0

	wasmSig := golang_ed25519.Sign(privKey, wasmDigest)
	signerBytes := golang_ed25519.PublicKey(privKey[32:])

	wasmSigHex := hex.EncodeToString(append(
		append(wasmDigest[:], ED25519ID),
		append(signerBytes[:], wasmSig[:]...)...,
	))

	require.Equal(t, originalSigHex, wasmSigHex)
}
