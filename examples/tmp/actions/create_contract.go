package actions

import (
	"context"
	"encoding/binary"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/storage"
	"github.com/ava-labs/hypersdk/state"

	mconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
)

var _ chain.Action = (*CreateContract)(nil)

type CreateContract struct {
	Bytecode      []byte
	Discriminator uint16
}

func (*CreateContract) GetTypeID() uint8 {
	return mconsts.CreateContractID
}

func (cc *CreateContract) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	contractAddress := storage.GenerateContractAddress(actor, cc.Discriminator)

	return state.Keys{
		string(storage.ContractBytecodeKey(contractAddress)): state.All,
	}
}

func (*CreateContract) StateKeysMaxChunks() []uint16 {
	return []uint16{storage.ContractBytecodeChunks}
}

func (cc *CreateContract) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	addr, err := storage.CreateContract(ctx, mu, actor, cc.Bytecode, cc.Discriminator)
	if err != nil {
		return nil, err // FIXME: Consider defining distinct errors in outputs.go for better clarity
	}

	addrString := codec.MustAddressBech32(mconsts.HRP, addr)

	return [][]byte{
		[]byte(addrString),
	}, nil
}

func (*CreateContract) ComputeUnits(chain.Rules) uint64 {
	return CreateContractComputeUnits
}

func (cc *CreateContract) Size() int {
	return len(cc.Bytecode) + consts.Uint8Len
}

func (cc *CreateContract) Marshal(p *codec.Packer) {
	p.PackBytes(cc.Bytecode)

	discriminatorBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(discriminatorBytes, cc.Discriminator)
	p.PackBytes(discriminatorBytes)
}

func UnmarshalCreateContract(p *codec.Packer) (chain.Action, error) {
	var action CreateContract
	p.UnpackBytes(-1, false, &action.Bytecode)

	var discriminatorBytes []byte = make([]byte, 2)
	p.UnpackBytes(2, false, &discriminatorBytes)
	action.Discriminator = binary.BigEndian.Uint16(discriminatorBytes)

	return &action, nil
}

func (*CreateContract) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
