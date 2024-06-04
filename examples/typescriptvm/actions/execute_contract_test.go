package actions

import (
	"bytes"
	"testing"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/ava-labs/hypersdk/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteContractSerialization(t *testing.T) {
	testAddress := codec.EmptyAddress
	testAddress[0] = 0x01
	testAddress[1] = 0x02
	testAddress[2] = 0x03
	testAddress[codec.AddressLen-1] = 0x04

	tests := []struct {
		name   string
		action ExecuteContract
	}{
		{
			name: "Empty fields",
			action: ExecuteContract{
				ContractAddress: codec.EmptyAddress,
				Payload:         nil,
				Keys:            nil,
			},
		},
		{
			name: "Only contract address",
			action: ExecuteContract{
				ContractAddress: testAddress,
				Payload:         nil,
				Keys:            nil,
			},
		},
		{
			name: "Only payload",
			action: ExecuteContract{
				ContractAddress: codec.EmptyAddress,
				Payload:         []byte("payload"),
				Keys:            nil,
			},
		},
		{
			name: "Only keys",
			action: ExecuteContract{
				ContractAddress: codec.EmptyAddress,
				Payload:         nil,
				Keys: map[runtime.KeyPostfix]state.Permissions{
					{0x01, 0x02, 0x03, 0x04}: state.Read,
				},
			},
		},
		{
			name: "All fields filled",
			action: ExecuteContract{
				ContractAddress: testAddress,
				Payload:         []byte("payload"),
				Keys: map[runtime.KeyPostfix]state.Permissions{
					{0x01, 0x02, 0x03, 0x04}: state.Read | state.Write,
					{0x05, 0x06, 0x07, 0x08}: state.Write,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			packer := codec.NewWriter(0, 99999)

			tt.action.Marshal(packer)
			if packer.Err() != nil {
				t.Fatalf("Marshal failed: %v", packer.Err())
			}

			n, err := buf.Write(packer.Bytes())
			require.NoError(t, err)
			require.Equal(t, len(packer.Bytes()), n)

			unpacker := codec.NewReader(buf.Bytes(), len(buf.Bytes()))
			unmarshalledAction, err := UnmarshalExecuteContract(unpacker)
			if tt.action.ContractAddress == codec.EmptyAddress {
				assert.Error(t, err)
				return
			} else if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			require.NotNil(t, unmarshalledAction)
			unmarshalledEC := unmarshalledAction.(*ExecuteContract)

			require.Equal(t, tt.action.ContractAddress, unmarshalledEC.ContractAddress, "ContractAddress mismatch")
			if tt.action.Payload == nil {
				require.Subset(t, [][]byte{tt.action.Payload, []byte{}}, [][]byte{unmarshalledEC.Payload}, "Payload mismatch")
			} else {
				require.Equal(t, tt.action.Payload, unmarshalledEC.Payload, "Payload mismatch")
			}
			require.Equal(t, len(tt.action.Keys), len(unmarshalledEC.Keys), "Keys length mismatch")
			for k, v := range tt.action.Keys {
				require.Equal(t, v, unmarshalledEC.Keys[k], "Permissions mismatch for key %v", k)
			}
		})
	}
}
