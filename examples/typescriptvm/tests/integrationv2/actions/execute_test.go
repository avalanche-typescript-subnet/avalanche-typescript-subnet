package integration_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	lconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/tests/integrationv2/runtime/assets"
	"github.com/ava-labs/hypersdk/state"
	"github.com/stretchr/testify/require"
)

//go:embed assets/simple_counter_copy.wasm
var testWasmBytes []byte

func TestRPCEcho(t *testing.T) {
	prep := prepare(t)

	contractAddrString := deployTestContractHelper(t, prep)

	//echo test
	callResult, err := prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, []byte{assets.CONTRACT_ACTION_ECHO, 70}, prep.addrStr)
	require.NoError(t, err)
	require.Equal(t, []byte("70"), callResult.Result)
}

func TestExecuteIncrement(t *testing.T) {
	prep := prepare(t)

	contractAddrString := deployTestContractHelper(t, prep)
	contractAddr, err := codec.ParseAddressBech32(lconsts.HRP, contractAddrString)
	require.NoError(t, err)

	//should be zero at first call
	callResult, err := prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, []byte{assets.CONTRACT_ACTION_READ}, prep.addrStr)
	require.NoError(t, err)
	require.Equal(t, []byte("0"), callResult.Result)

	parser, err := prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err := prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.ExecuteContract{
			ContractAddress: contractAddr,
			Payload:         []byte{assets.CONTRACT_ACTION_INCREMENT, 0x7},
			Keys: map[runtime.KeyPostfix]state.Permissions{
				runtime.KeyPostfix([]byte{0, 0, 0, 0xa}): state.All,
			},
			ComputeUnitsToSpend: 4,
		}},
		prep.factory,
	)
	require.NoError(t, err)
	require.NoError(t, submit(context.Background()))

	results := prep.expectBlk(t, prep.instance)(false)
	require.Len(t, results, 1)
	require.True(t, results[0].Success, string(results[0].Error))

	//check again
	callResult, err = prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, []byte{assets.CONTRACT_ACTION_READ}, prep.addrStr)
	require.NoError(t, err)
	require.Equal(t, []byte("7"), callResult.Result)
}
