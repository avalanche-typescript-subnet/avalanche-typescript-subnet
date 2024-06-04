package integration_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/tests/integrationv2/runtime/assets"
	"github.com/stretchr/testify/require"
)

//go:embed assets/simple_counter_copy.wasm
var testWasmBytes []byte

func TestRPCEcho(t *testing.T) {
	discriminator := 111

	//send CreateContract tx

	prep := prepare(t)

	parser, err := prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err := prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.CreateContract{
			Bytecode:      testWasmBytes,
			Discriminator: uint16(discriminator),
		}},
		prep.factory,
	)
	require.NoError(t, err)
	require.NoError(t, submit(context.Background()))

	results := prep.expectBlk(t, prep.instance)(false)
	require.Len(t, results, 1)
	require.True(t, results[0].Success)

	contractAddrString := string(results[0].Outputs[0][0])

	//contract deployed, echo test
	callResult, err := prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, []byte{assets.CONTRACT_ACTION_ECHO, 70}, prep.addrStr)
	require.NoError(t, err)
	require.Equal(t, []byte("70"), callResult.Result)
}
