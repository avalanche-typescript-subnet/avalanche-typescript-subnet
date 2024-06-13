package integration_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	lconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
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

	//execute increment only to figure out keys
	incrementPayload := []byte{assets.CONTRACT_ACTION_INCREMENT, 0x7}

	callResult, err = prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, incrementPayload, prep.addrStr)
	require.NoError(t, err)

	//now execute increment in transaction
	parser, err := prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err := prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.ExecuteContract{
			ContractAddress:     contractAddr,
			Payload:             incrementPayload,
			Keys:                callResult.Keys,
			ComputeUnitsToSpend: callResult.ComputeUnitsSpent,
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

func TestExecuteManyReadsAndWrites(t *testing.T) {
	slotsToWrite := byte(10)

	prep := prepare(t)

	contractAddrString := deployTestContractHelper(t, prep)
	contractAddr, err := codec.ParseAddressBech32(lconsts.HRP, contractAddrString)
	require.NoError(t, err)

	//execute WRITE_MANY_SLOTS to write slotsToWrite slots
	writePayload := []byte{assets.CONTRACT_ACTION_WRITE_MANY_SLOTS, slotsToWrite}

	callResult, err := prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, writePayload, prep.addrStr)
	require.NoError(t, err)

	//now execute increment in transaction
	parser, err := prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err := prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.ExecuteContract{
			ContractAddress:     contractAddr,
			Payload:             writePayload,
			Keys:                callResult.Keys,
			ComputeUnitsToSpend: callResult.ComputeUnitsSpent,
		}},
		prep.factory,
	)
	require.NoError(t, err)
	require.NoError(t, submit(context.Background()))

	results := prep.expectBlk(t, prep.instance)(false)
	require.Len(t, results, 1)
	require.True(t, results[0].Success, string(results[0].Error))

	//execute a bunch of reads in a transaction
	readPayload := []byte{assets.CONTRACT_ACTION_READ_MANY_SLOTS, slotsToWrite}

	callResult, err = prep.instance.lcli.ExecuteContract(context.Background(), contractAddrString, readPayload, prep.addrStr)
	require.NoError(t, err)

	//now execute increment in transaction
	parser, err = prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err = prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.ExecuteContract{
			ContractAddress:     contractAddr,
			Payload:             readPayload,
			Keys:                callResult.Keys,
			ComputeUnitsToSpend: callResult.ComputeUnitsSpent,
		}},
		prep.factory,
	)
	require.NoError(t, err)
	require.NoError(t, submit(context.Background()))

	results = prep.expectBlk(t, prep.instance)(false)
	require.Len(t, results, 1)
	require.True(t, results[0].Success, string(results[0].Error))
}
