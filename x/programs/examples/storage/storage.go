// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/state"
)

const (
	programPrefix = 0x0
)

// ProgramPrefixKey returns a properly formatted key
// for storing a value at [id][key].
func ProgramPrefixKey(id []byte, key []byte) (k []byte) {
	k = make([]byte, 1+consts.IDLen+len(key))
	k[0] = programPrefix
	copy(k[1:1+consts.IDLen], id[:])
	copy(k[1+consts.IDLen+1:], key[:])
	return
}

// ProgramKey returns the key used to store the program bytes at [id].
func ProgramKey(id ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = programPrefix
	copy(k[1:], id[:])
	return
}

// GetProgram returns the programBytes stored at [programID].
func GetProgram(
	ctx context.Context,
	db state.Immutable,
	programID ids.ID,
) (
	[]byte, // program bytes
	bool, // exists
	error,
) {
	k := ProgramKey(programID)
	v, err := db.GetValue(ctx, k)
	if errors.Is(err, database.ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// SetProgram stores [program] at [programID]
func SetProgram(
	ctx context.Context,
	mu state.Mutable,
	programID ids.ID,
	program []byte,
) error {
	k := ProgramKey(programID)
	return mu.Insert(ctx, k, program)
}
