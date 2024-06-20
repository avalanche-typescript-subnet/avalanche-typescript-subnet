package actions

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"time"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"

	mconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/storage"
)

var _ chain.Action = (*ExecuteContract)(nil)

type ExecuteContract struct {
	ContractAddress     codec.Address            `json:"contractAddress"`
	Payload             []byte                   `json:"payload"`
	FunctionName        string                   `json:"functionName"`
	Keys                StateKeysWithPermissions `json:"stateKeys"`
	ComputeUnitsToSpend uint64                   `json:"computeUnitsToSpend"`
}

func (*ExecuteContract) GetTypeID() uint8 {
	return mconsts.ExecuteContractID
}

func (ec *ExecuteContract) StateKeys(actor codec.Address, _ ids.ID) state.Keys {
	keys := make(state.Keys, len(ec.Keys))
	for k, v := range ec.Keys {
		keys[string(storage.ContractStateKey(ec.ContractAddress, []byte(k)))] = v
	}
	keys[string(storage.ContractBytecodeKey(ec.ContractAddress))] = state.Read

	//debug
	for k, v := range keys {
		fmt.Printf("StateKeys: key: %x, val: %d\n", []byte(k), v)
	}

	return keys
}

// FIXME: This code was copied from hypersdk/examples/typescriptvm/actions/evm_call.go in the evm-call-experiment branch. It needs testing.
func (ec *ExecuteContract) StateKeysMaxChunks() []uint16 {
	output := make([]uint16, 0, len(ec.Keys))
	for key := range ec.Keys {
		keyBytes := []byte(key)
		maxChunks := binary.BigEndian.Uint16(keyBytes[len(keyBytes)-2:])
		output = append(output, maxChunks)
	}
	return output
}

func (ec *ExecuteContract) Execute(
	ctx context.Context,
	_ chain.Rules,
	mu state.Mutable,
	_ int64,
	actor codec.Address,
	_ ids.ID,
) ([][]byte, error) {
	bytecode, err := storage.GetContractBytecode(ctx, mu, ec.ContractAddress)
	if err != nil {
		return nil, err
	}

	precachedValues := make(map[string][]byte)

	for keyPostfix, permissions := range ec.Keys {
		if permissions&state.Read == 0 {
			continue
		}

		val, err := storage.GetContractStateValue(ctx, mu, ec.ContractAddress, keyPostfix)
		if err != nil {
			return nil, fmt.Errorf("failed to get contract state value: %w", err)
		}
		precachedValues[keyPostfix] = val
	}

	var fixedStateProvider runtime.StateProvider = func(key string) ([]byte, error) {
		return precachedValues[string(key)], nil
	}

	params := runtime.JavyExecParams{ // FIXME:move limits to config
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 10,
		MaxMemory:     1024 * 1024 * 10,
		Bytecode:      &bytecode,
		StateProvider: fixedStateProvider,
		Payload:       ec.Payload,
		Actor:         actor[:],
		FunctionName:  ec.FunctionName,
	}

	res, err := runtime.NewJavyExec().Execute(params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute contract: %w", err)
	}

	if res.Result.Success != true {
		return nil, fmt.Errorf("contract execution failed: %s", res.Result.Error)
	}

	err = storage.UpdateContractStateFields(ctx, mu, ec.ContractAddress, res.Result.UpdatedKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to update contract state: %w", err)
	}

	computeUnitsSpent := res.FuelConsumed / 1_000_000 // TODO: move to consts
	computeUnitsSpent = max(ExecuteContractMinComputeUnits, computeUnitsSpent)

	if computeUnitsSpent != ec.ComputeUnitsToSpend {
		return nil, fmt.Errorf("compute units spent (%d) does not equal the compute units to spend (%d)", computeUnitsSpent, ec.ComputeUnitsToSpend)
	}

	return [][]byte{res.Result.Result}, nil

}

func (ec *ExecuteContract) ComputeUnits(chain.Rules) uint64 {
	return ec.ComputeUnitsToSpend
}

func (ec *ExecuteContract) Size() int {
	return len(ec.ContractAddress) + len(ec.Payload) + len(ec.Keys)*(4+1)
}

func (ec *ExecuteContract) Marshal(p *codec.Packer) {
	p.PackAddress(ec.ContractAddress)
	packBytesOrNull(p, ec.Payload)
	p.PackString(ec.FunctionName)
	marshalKeys(ec.Keys, p)
	p.PackUint64(ec.ComputeUnitsToSpend)
}

func UnmarshalExecuteContract(p *codec.Packer) (chain.Action, error) {
	var err error

	var executeContract ExecuteContract

	p.UnpackAddress(&executeContract.ContractAddress)
	err = unmarshalBytesOrNull(p, &executeContract.Payload)
	if err != nil {
		return nil, err
	}

	executeContract.FunctionName = p.UnpackString(false)

	executeContract.Keys, err = unmarshalKeys(p)
	if err != nil {
		return nil, err
	}

	executeContract.ComputeUnitsToSpend = p.UnpackUint64(false)

	return &executeContract, nil
}

func (*ExecuteContract) ValidRange(chain.Rules) (int64, int64) {
	// Returning -1, -1 means that the action is always valid.
	return -1, -1
}
func marshalKeys(keys map[string]state.Permissions, p *codec.Packer) {
	p.PackInt(len(keys))
	keysOrdered := make([][]byte, 0, len(keys))
	for k := range keys {
		keysOrdered = append(keysOrdered, []byte(k))
	}
	sort.Slice(keysOrdered, func(i, j int) bool {
		return bytes.Compare(keysOrdered[i][:], keysOrdered[j][:]) < 0
	})

	for _, k := range keysOrdered { // Iterate over the keys in sorted order
		p.PackBytes(k)                    // Serialize the 4-byte KeyPostfix
		p.PackByte(byte(keys[string(k)])) // Serialize the permissions associated with the key
	}
}

func unmarshalKeys(p *codec.Packer) (StateKeysWithPermissions, error) {
	numKeys := p.UnpackInt(false)
	keys := make(StateKeysWithPermissions, numKeys)
	for i := 0; i < numKeys; i++ {
		var keyPostfix []byte
		p.UnpackBytes(10000, false, &keyPostfix) // Deserialize the 4-byte KeyPostfix
		perm := state.Permissions(p.UnpackByte())
		keys[string(keyPostfix)] = perm
	}
	return keys, p.Err()
}

func packBytesOrNull(p *codec.Packer, b []byte) {
	flag := b != nil
	p.PackBool(flag)
	if flag {
		p.PackBytes(b)
	}
}

func unmarshalBytesOrNull(p *codec.Packer, field *[]byte) error {
	flag := p.UnpackBool()
	if flag {
		p.UnpackBytes(-1, false, field)
	} else {
		*field = nil
	}
	return p.Err()
}

type StateKeysWithPermissions map[string]state.Permissions
