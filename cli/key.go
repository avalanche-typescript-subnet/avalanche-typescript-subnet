// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cli

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/utils"
	"golang.org/x/exp/maps"
)

func (h *Handler) SetKey(lookupBalance func(int, string, string, uint32, ids.ID) error) error {
	keys, err := h.GetKeys()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		utils.Outf("{{red}}no stored keys{{/}}\n")
		return nil
	}
	chainID, uris, err := h.GetDefaultChain(true)
	if err != nil {
		return err
	}
	if len(uris) == 0 {
		utils.Outf("{{red}}no available chains{{/}}\n")
		return nil
	}
	uriName := maps.Keys(uris)[0]
	rcli := rpc.NewJSONRPCClient(uris[uriName])
	networkID, _, _, err := rcli.Network(context.TODO())
	if err != nil {
		return err
	}
	utils.Outf("{{cyan}}stored keys:{{/}} %d\n", len(keys))
	for i := 0; i < len(keys); i++ {
		if err := lookupBalance(i, h.c.Address(keys[i].Address), uris[uriName], networkID, chainID); err != nil {
			return err
		}
	}

	// Select key
	keyIndex, err := h.PromptChoice("set default key", len(keys))
	if err != nil {
		return err
	}
	key := keys[keyIndex]
	return h.StoreDefaultKey(key.Address)
}

func (h *Handler) Balance(checkAllChains bool, promptAsset bool, printBalance func(codec.Address, string, uint32, ids.ID, ids.ID) error) error {
	addr, _, err := h.GetDefaultKey(true)
	if err != nil {
		return err
	}
	chainID, uris, err := h.GetDefaultChain(true)
	if err != nil {
		return err
	}
	var assetID ids.ID
	if promptAsset {
		assetID, err = h.PromptAsset("assetID", true)
		if err != nil {
			return err
		}
	}

	uriNames := maps.Keys(uris)
	max := len(uriNames)
	if !checkAllChains {
		max = 1
	}
	for _, uri := range uriNames[:max] {
		utils.Outf("{{yellow}}uri:{{/}} %s\n", uri)
		rcli := rpc.NewJSONRPCClient(uri)
		networkID, _, _, err := rcli.Network(context.TODO())
		if err != nil {
			return err
		}
		if err := printBalance(addr, uri, networkID, chainID, assetID); err != nil {
			return err
		}
	}
	return nil
}
