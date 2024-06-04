package storage

import (
	"context"
	"encoding/binary"
	"errors"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	mconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
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
	discriminator uint16,
) (codec.Address, error) {
	contractAddress := GenerateContractAddress(addr, discriminator)
	bytecodeKey := ContractBytecodeKey(contractAddress)

	_, err := mu.GetValue(ctx, bytecodeKey)
	if err == nil {
		return codec.EmptyAddress, errors.New("contract already exists")
	} else if !errors.Is(err, database.ErrNotFound) {
		return codec.EmptyAddress, err
	}

	err = mu.Insert(ctx, bytecodeKey, bytecode)
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

func ContractStateKey(contractAddr codec.Address, postfix [4]byte) []byte {
	return append(append([]byte{contractStatePrefix}, contractAddr[:]...), postfix[:]...)
}

func GetContractStateProviderFromState(
	ctx context.Context,
	f ReadState,
	addr codec.Address,
) runtime.StateProvider {
	return func(postfix [4]byte) ([]byte, error) { //FIXME: would be more efficient to batch get all state keys
		k := ContractStateKey(addr, postfix)
		values, errs := f(ctx, [][]byte{k})

		//this allows  contracts to read non-existing state as empty bytes
		if errors.Is(errs[0], database.ErrNotFound) {
			return []byte{}, nil
		}
		return values[0], errs[0]
	}
}
