// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package v2

import "C"
import (
	"context"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/x/programs/program"
	"github.com/bytecodealliance/wasmtime-go/v14"
	"github.com/near/borsh-go"
)

const (
	AllocName  = "alloc"
	MemoryName = "memory"
)

type CallInfo struct {
	State           state.Mutable
	Actor           ids.ID
	StateAccessList StateAccessList
	Account         ids.ID
	ProgramID       ids.ID
	Fuel            uint64
	FunctionName    string
	Params          []byte
	result          []byte
}

type Program struct {
	module    *wasmtime.Module
	programID ids.ID
}

type ProgramInstance struct {
	*Program
	inst  *wasmtime.Instance
	store *wasmtime.Store
}

func newProgram(engine *wasmtime.Engine, programID ids.ID, programBytes []byte) (*Program, error) {
	module, err := wasmtime.NewModule(engine, programBytes)
	if err != nil {
		return nil, err
	}
	return &Program{module: module, programID: programID}, nil
}

func (p *ProgramInstance) call(_ context.Context, callInfo *CallInfo) ([]byte, error) {
	if err := p.store.AddFuel(callInfo.Fuel); err != nil {
		return nil, err
	}

	// create the program context
	programCtx := program.Context{ProgramID: callInfo.ProgramID}
	programCtxBytes, err := borsh.Serialize(programCtx)
	if err != nil {
		return nil, err
	}

	//copy context into store linear memory
	ctxOffset, err := p.setParam(programCtxBytes)
	if err != nil {
		return nil, err
	}
	if callInfo.Params == nil {
		_, err = p.inst.GetFunc(p.store, callInfo.FunctionName).Call(p.store, ctxOffset)
	} else {
		// if params exist, copy them into linear memory too
		paramsOffset, err := p.setParam(programCtxBytes)
		if err != nil {
			return nil, err
		}
		_, err = p.inst.GetFunc(p.store, callInfo.FunctionName).Call(p.store, ctxOffset, paramsOffset)
	}

	return callInfo.result, err
}

func (p *ProgramInstance) setParam(data []byte) (int32, error) {
	allocFn := p.inst.GetExport(p.store, AllocName).Func()
	programMemory := p.inst.GetExport(p.store, MemoryName).Memory()
	dataOffsetIntf, err := allocFn.Call(p.store, int32(len(data)))
	if err != nil {
		return 0, wasmtime.NewTrap(err.Error())
	}
	dataOffset := dataOffsetIntf.(int32)
	linearMem := programMemory.UnsafeData(p.store)
	copy(linearMem[dataOffset:], data)
	return dataOffset, nil
}