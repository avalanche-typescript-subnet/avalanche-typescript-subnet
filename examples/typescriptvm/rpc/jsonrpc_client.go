// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"context"
	"strings"

	"github.com/ava-labs/avalanchego/ids"

	_ "github.com/ava-labs/hypersdk/examples/typescriptvm/registry" // ensure registry populated
	"github.com/ava-labs/hypersdk/state"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/genesis"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/storage"
	"github.com/ava-labs/hypersdk/requester"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
)

type JSONRPCClient struct {
	requester *requester.EndpointRequester

	networkID uint32
	chainID   ids.ID
	g         *genesis.Genesis
}

// New creates a new client object.
func NewJSONRPCClient(uri string, networkID uint32, chainID ids.ID) *JSONRPCClient {
	uri = strings.TrimSuffix(uri, "/")
	uri += JSONRPCEndpoint
	req := requester.New(uri, consts.Name)
	return &JSONRPCClient{req, networkID, chainID, nil}
}

func (cli *JSONRPCClient) Genesis(ctx context.Context) (*genesis.Genesis, error) {
	if cli.g != nil {
		return cli.g, nil
	}

	resp := new(GenesisReply)
	err := cli.requester.SendRequest(
		ctx,
		"genesis",
		nil,
		resp,
	)
	if err != nil {
		return nil, err
	}
	cli.g = resp.Genesis
	return resp.Genesis, nil
}

func (cli *JSONRPCClient) Tx(ctx context.Context, id ids.ID) (bool, bool, int64, uint64, error) {
	resp := new(TxReply)
	err := cli.requester.SendRequest(
		ctx,
		"tx",
		&TxArgs{TxID: id},
		resp,
	)
	switch {
	// We use string parsing here because the JSON-RPC library we use may not
	// allows us to perform errors.Is.
	case err != nil && strings.Contains(err.Error(), ErrTxNotFound.Error()):
		return false, false, -1, 0, nil
	case err != nil:
		return false, false, -1, 0, err
	}
	return true, resp.Success, resp.Timestamp, resp.Fee, nil
}

func (cli *JSONRPCClient) Balance(ctx context.Context, addr string) (uint64, error) {
	resp := new(BalanceReply)
	err := cli.requester.SendRequest(
		ctx,
		"balance",
		&BalanceArgs{
			Address: addr,
		},
		resp,
	)
	return resp.Amount, err
}

func (cli *JSONRPCClient) WaitForBalance(
	ctx context.Context,
	addr string,
	min uint64,
) error {
	return rpc.Wait(ctx, func(ctx context.Context) (bool, error) {
		balance, err := cli.Balance(ctx, addr)
		if err != nil {
			return false, err
		}
		shouldExit := balance >= min
		if !shouldExit {
			utils.Outf(
				"{{yellow}}waiting for %s balance: %s{{/}}\n",
				utils.FormatBalance(min, consts.Decimals),
				addr,
			)
		}
		return shouldExit, nil
	})
}

func (cli *JSONRPCClient) WaitForTransaction(ctx context.Context, txID ids.ID) (bool, uint64, error) {
	var success bool
	var fee uint64
	if err := rpc.Wait(ctx, func(ctx context.Context) (bool, error) {
		found, isuccess, _, ifee, err := cli.Tx(ctx, txID)
		if err != nil {
			return false, err
		}
		success = isuccess
		fee = ifee
		return found, nil
	}); err != nil {
		return false, 0, err
	}
	return success, fee, nil
}

var _ chain.Parser = (*Parser)(nil)

type Parser struct {
	networkID uint32
	chainID   ids.ID
	genesis   *genesis.Genesis
}

func (p *Parser) ChainID() ids.ID {
	return p.chainID
}

func (p *Parser) Rules(t int64) chain.Rules {
	return p.genesis.Rules(t, p.networkID, p.chainID)
}

func (*Parser) Registry() (chain.ActionRegistry, chain.AuthRegistry) {
	return consts.ActionRegistry, consts.AuthRegistry
}

func (*Parser) StateManager() chain.StateManager {
	return &storage.StateManager{}
}

func (cli *JSONRPCClient) Parser(ctx context.Context) (chain.Parser, error) {
	g, err := cli.Genesis(ctx)
	if err != nil {
		return nil, err
	}
	return &Parser{cli.networkID, cli.chainID, g}, nil
}

func (cli *JSONRPCClient) ContractBytecode(ctx context.Context, addr string) ([]byte, error) {
	resp := new(ContractBytecodeReply)
	err := cli.requester.SendRequest(
		ctx,
		"contractBytecode",
		&ContractBytecodeArgs{
			Address: addr,
		},
		resp,
	)
	return resp.Bytecode, err
}

type ExecuteContractClientReply struct {
	DebugLog          string
	Result            []byte
	Success           bool
	Error             string
	Keys              map[string]state.Permissions
	ComputeUnitsSpent uint64
}

func (cli *JSONRPCClient) ExecuteContract(ctx context.Context, addr string, funcName string, input []byte, actor string) (ExecuteContractClientReply, error) {
	originalResp := new(ExecuteContractReply)
	err := cli.requester.SendRequest(
		ctx,
		"executeContract",
		&ExecuteContractArgs{
			ContractAddress: addr,
			Payload:         input,
			FunctionName:    funcName,
			Actor:           actor,
		},
		originalResp,
	)

	resp := new(ExecuteContractClientReply)
	resp.DebugLog = originalResp.DebugLog
	resp.Result = originalResp.Result
	resp.Success = originalResp.Success
	resp.Error = originalResp.Error
	resp.ComputeUnitsSpent = originalResp.ComputeUnitsSpent

	resp.Keys = make(map[string]state.Permissions)

	for _, key := range originalResp.ReadKeys {
		resp.Keys[string(key)] = state.Read
	}
	for _, key := range originalResp.UpdatedKeys {
		if _, hadRead := (resp.Keys)[string(key)]; !hadRead {
			resp.Keys[string(key)] = state.Write | state.Allocate
		} else {
			resp.Keys[string(key)] = state.Write | state.Allocate | state.Read
		}
	}

	return *resp, err
}
