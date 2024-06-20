// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package rpc

import (
	"net/http"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/genesis"
	"github.com/ava-labs/hypersdk/fees"
)

type JSONRPCServer struct {
	c Controller
}

func NewJSONRPCServer(c Controller) *JSONRPCServer {
	return &JSONRPCServer{c}
}

type GenesisReply struct {
	Genesis *genesis.Genesis `json:"genesis"`
}

func (j *JSONRPCServer) Genesis(_ *http.Request, _ *struct{}, reply *GenesisReply) (err error) {
	reply.Genesis = j.c.Genesis()
	return nil
}

type TxArgs struct {
	TxID ids.ID `json:"txId"`
}

type TxReply struct {
	Timestamp int64           `json:"timestamp"`
	Success   bool            `json:"success"`
	Units     fees.Dimensions `json:"units"`
	Fee       uint64          `json:"fee"`
}

func (j *JSONRPCServer) Tx(req *http.Request, args *TxArgs, reply *TxReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Tx")
	defer span.End()

	found, t, success, units, fee, err := j.c.GetTransaction(ctx, args.TxID)
	if err != nil {
		return err
	}
	if !found {
		return ErrTxNotFound
	}
	reply.Timestamp = t
	reply.Success = success
	reply.Units = units
	reply.Fee = fee
	return nil
}

type BalanceArgs struct {
	Address string `json:"address"`
}

type BalanceReply struct {
	Amount uint64 `json:"amount"`
}

func (j *JSONRPCServer) Balance(req *http.Request, args *BalanceArgs, reply *BalanceReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.Balance")
	defer span.End()

	addr, err := codec.ParseAddressBech32(consts.HRP, args.Address)
	if err != nil {
		return err
	}
	balance, err := j.c.GetBalanceFromState(ctx, addr)
	if err != nil {
		return err
	}
	reply.Amount = balance
	return err
}

type ContractBytecodeArgs struct {
	Address string `json:"address"`
}

type ContractBytecodeReply struct {
	Bytecode []byte `json:"bytecode"`
}

func (j *JSONRPCServer) ContractBytecode(req *http.Request, args *ContractBytecodeArgs, reply *ContractBytecodeReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.ContractBytecode")
	defer span.End()

	addr, err := codec.ParseAddressBech32(consts.HRP, args.Address)
	if err != nil {
		return err
	}
	bytecode, err := j.c.GetContractBytecodeFromState(ctx, addr)
	if err != nil {
		return err
	}
	reply.Bytecode = bytecode
	return err
}

type ExecuteContractArgs struct {
	ContractAddress string `json:"contractAddress"`
	FunctionName    string `json:"functionName"`
	Payload         []byte `json:"payload"`
	Actor           string `json:"actor"`
}

type ExecuteContractReply struct {
	DebugLog          string   `json:"debugLog"`
	Result            []byte   `json:"result"`
	Success           bool     `json:"success"`
	Error             string   `json:"error"`
	UpdatedKeys       [][]byte `json:"updatedKeys"`
	ReadKeys          [][]byte `json:"readKeys"`
	ComputeUnitsSpent uint64   `json:"computeUnitsSpent"`
}

func (j *JSONRPCServer) ExecuteContract(req *http.Request, args *ExecuteContractArgs, reply *ExecuteContractReply) error {
	ctx, span := j.c.Tracer().Start(req.Context(), "Server.ExecuteContract")
	defer span.End()

	contractAddr, err := codec.ParseAddressBech32(consts.HRP, args.ContractAddress)
	if err != nil {
		return err
	}

	actorAddr, err := codec.ParseAddressBech32(consts.HRP, args.Actor)
	if err != nil {
		return err
	}

	res, err := j.c.ExecuteContractOnState(ctx, contractAddr, actorAddr, args.Payload, args.FunctionName)
	if err != nil {
		return err
	}

	reply.DebugLog = string(res.DebugLog)
	reply.Result = res.Result.Result
	reply.Success = res.Result.Success
	reply.Error = res.Result.Error

	// Convert each [4]byte to KeyPostfix and assign to reply.ReadKeys
	reply.ReadKeys = make([][]byte, 0, len(res.Result.ReadKeys))
	for _, key := range res.Result.ReadKeys {
		reply.ReadKeys = append(reply.ReadKeys, []byte(key))
	}

	reply.UpdatedKeys = make([][]byte, 0, len(res.Result.UpdatedKeys))
	for key := range res.Result.UpdatedKeys {
		reply.UpdatedKeys = append(reply.UpdatedKeys, []byte(key))
	}

	reply.ComputeUnitsSpent = res.FuelConsumed / 1_000_000 // TODO: move to consts
	reply.ComputeUnitsSpent = max(actions.ExecuteContractMinComputeUnits, reply.ComputeUnitsSpent)

	return nil
}
