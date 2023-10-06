// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package examples

import (
	"context"
	"github.com/ava-labs/avalanchego/utils/logging"

	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/x/programs/runtime"
	"github.com/ava-labs/hypersdk/x/programs/utils"

	"go.uber.org/zap"
)

func NewNFT(log logging.Logger, programBytes []byte, maxFee uint64, costMap map[string]uint64, m Metadata) *NFT {
	return &NFT{
		log:          log,
		programBytes: programBytes,
		metadata:     m,
		maxFee:       maxFee,
		costMap:      costMap,
	}
}

type NFT struct {
	log          logging.Logger
	programBytes []byte
	metadata     Metadata
	// metering
	maxFee  uint64
	costMap map[string]uint64
}

type Metadata struct {
	Name   string
	Symbol string
	URI    string
}

func (t *NFT) Run(ctx context.Context) error {
	meter := runtime.NewMeter(t.log, t.maxFee, t.costMap)
	db := utils.NewTestDB()
	store := newProgramStorage(db)

	runtime := runtime.New(t.log, meter, store)
	contractId, err := runtime.Create(ctx, t.programBytes)
	if err != nil {
		return err
	}

	t.log.Debug("initial cost",
		zap.Int("gas", 0),
	)

	// Generate keys for Alice
	alicePtr, _, err := nftKeyPtr(ctx, runtime)
	if err != nil {
		return err
	}

	_, err = runtime.Call(ctx, "init", contractId)
	if err != nil {
		return err
	}
	t.log.Debug("init called")

	// mint 1 token, send to alice
	_, err = runtime.Call(ctx, "mint", contractId, alicePtr)
	if err != nil {
		return err
	}
	t.log.Debug("minted")

	return nil
}

func nftKeyPtr(ctx context.Context, runtime runtime.Runtime) (uint64, ed25519.PublicKey, error) {
	priv, err := ed25519.GeneratePrivateKey()
	if err != nil {
		return 0, ed25519.EmptyPublicKey, err
	}

	pk := priv.PublicKey()
	ptr, err := runtime.WriteGuestBuffer(ctx, pk[:])
	return ptr, pk, err
}
