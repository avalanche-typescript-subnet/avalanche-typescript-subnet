// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package controller

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/trace"
	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/genesis"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/storage"
	"github.com/ava-labs/hypersdk/fees"
)

func (c *Controller) Genesis() *genesis.Genesis {
	return c.genesis
}

func (c *Controller) Logger() logging.Logger {
	return c.inner.Logger()
}

func (c *Controller) Tracer() trace.Tracer {
	return c.inner.Tracer()
}

func (c *Controller) GetTransaction(
	ctx context.Context,
	txID ids.ID,
) (bool, int64, bool, fees.Dimensions, uint64, error) {
	return storage.GetTransaction(ctx, c.metaDB, txID)
}

func (c *Controller) GetBalanceFromState(
	ctx context.Context,
	acct codec.Address,
) (uint64, error) {
	return storage.GetBalanceFromState(ctx, c.inner.ReadState, acct)
}

func (c *Controller) GetContractBytecodeFromState(
	ctx context.Context,
	acct codec.Address,
) ([]byte, error) {
	return storage.GetContractBytecodeFromState(ctx, c.inner.ReadState, acct)
}

func (c *Controller) ExecuteContractOnState(
	ctx context.Context,
	contractAddress codec.Address,
	actor codec.Address,
	payload []byte,
	funcName string,
) (*runtime.JavyExecResult, error) {
	bytecode, err := c.GetContractBytecodeFromState(ctx, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract bytecode: %w", err)
	}
	if bytecode == nil {
		return nil, fmt.Errorf("contract %s has no bytecode", contractAddress)
	}

	provider := storage.GetContractStateProviderFromState(ctx, c.inner.ReadState, contractAddress)

	params := runtime.JavyExecParams{ // FIXME:move limits to config
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 10,
		MaxMemory:     1024 * 1024 * 10,
		Bytecode:      &bytecode,
		StateProvider: provider,
		Payload:       payload,
		Actor:         actor[:],
		FunctionName:  funcName,
	}

	var callback runtime.CallbackFunc = func(address []byte) ([]byte, error) {
		addrHex := string(address)

		addressBytes, err := hex.DecodeString(strings.TrimPrefix(addrHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("error decoding address %s: %v", addrHex, err)
		}

		newVal, err := params.StateProvider(string(addressBytes))
		if err != nil {
			return nil, fmt.Errorf("error retrieving state for address %x: %v", addressBytes, err)
		}
		return newVal, nil
	}

	return runtime.NewJavyExec().SetCallbackFunc(callback).Execute(params)
}
