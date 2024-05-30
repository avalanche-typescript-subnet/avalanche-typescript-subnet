package storage

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	mconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/state"
	"github.com/ava-labs/hypersdk/utils"
)

// [contractBytecodePrefix] + [address]
func ContractBytecodeKey(addr codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)
	k[0] = contractBytecodePrefix
	copy(k[1:], addr[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], ContractBytecodeChunks)
	return
}

// [contractStatePrefix] + [address]
func ContractStateKey(addr codec.Address) (k []byte) {
	k = make([]byte, 1+codec.AddressLen+consts.Uint16Len)
	k[0] = contractStatePrefix
	copy(k[1:], addr[:])
	binary.BigEndian.PutUint16(k[1+codec.AddressLen:], ContractStateChunks)
	return
}

func GenerateContractAddress(sender codec.Address, discriminator uint16) codec.Address {
	combinedBytes := make([]byte, 2+codec.AddressLen)
	copy(combinedBytes, sender[:])
	binary.BigEndian.PutUint16(combinedBytes[codec.AddressLen:], discriminator)
	id := utils.ToID(combinedBytes)

	return codec.CreateAddress(mconsts.SMARTCONTRACTID, id)
}

func CreateContract(
	ctx context.Context,
	mu state.Mutable,
	addr codec.Address,
	bytecode []byte,
	initialState []byte,
	discriminator uint16,
) (codec.Address, error) {
	contractAddress := GenerateContractAddress(addr, discriminator)
	bytecodeKey := ContractBytecodeKey(contractAddress)
	stateKey := ContractStateKey(contractAddress)

	_, err := mu.GetValue(ctx, bytecodeKey)
	if err == nil {
		return codec.EmptyAddress, errors.New("contract already exists")
	} else if !errors.Is(err, database.ErrNotFound) {
		return codec.EmptyAddress, err
	}

	_, err = mu.GetValue(ctx, stateKey)
	if err == nil {
		return codec.EmptyAddress, errors.New("contract already exists")
	} else if !errors.Is(err, database.ErrNotFound) {
		return codec.EmptyAddress, err
	}

	err = mu.Insert(ctx, bytecodeKey, bytecode)
	if err != nil {
		return codec.EmptyAddress, err
	}

	err = mu.Insert(ctx, stateKey, initialState)
	if err != nil {
		return codec.EmptyAddress, err
	}

	return contractAddress, nil
}

// Used to serve RPC queries
func GetContractBytecodeFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
) ([]byte, error) {
	k := ContractBytecodeKey(addr)
	values, errs := f(ctx, [][]byte{k})

	if errors.Is(errs[0], database.ErrNotFound) {
		return []byte{}, nil
	}

	return values[0], errs[0]
}

func GetContractStateFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
) ([]byte, error) {
	k := ContractStateKey(addr)
	values, errs := f(ctx, [][]byte{k})

	if errors.Is(errs[0], database.ErrNotFound) {
		return []byte{}, nil
	}

	return values[0], errs[0]
}
