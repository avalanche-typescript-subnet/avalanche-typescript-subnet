package actions

import (
	"bytes"
	"testing"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/state"
	"github.com/stretchr/testify/require"
)

func TestExecuteContractSerialization(t *testing.T) {
	testAddress := codec.Address{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}

	tests := []struct {
		name          string
		action        ExecuteContract
		errorExpected bool
	}{
		{
			name: "Empty address (error expected)",
			action: ExecuteContract{
				ContractAddress:     codec.EmptyAddress,
				Payload:             nil,
				Keys:                nil,
				ComputeUnitsToSpend: 0,
				FunctionName:        "",
			},
			errorExpected: true,
		},
		{
			name: "Only contract address",
			action: ExecuteContract{
				ContractAddress:     testAddress,
				Payload:             nil,
				Keys:                nil,
				ComputeUnitsToSpend: 0,
				FunctionName:        "",
			},
		},
		{
			name: "Payload",
			action: ExecuteContract{
				ContractAddress:     testAddress,
				Payload:             []byte("payload"),
				Keys:                nil,
				ComputeUnitsToSpend: 0,
				FunctionName:        "",
			},
		},
		{
			name: "Empty payload",
			action: ExecuteContract{
				ContractAddress:     testAddress,
				Payload:             []byte{},
				Keys:                nil,
				ComputeUnitsToSpend: 0,
				FunctionName:        "",
			},
		},
		{
			name: "Function name",
			action: ExecuteContract{
				ContractAddress:     testAddress,
				Payload:             nil,
				Keys:                nil,
				ComputeUnitsToSpend: 0,
				FunctionName:        "functionName",
			},
		},
		{
			name: "Compute units",
			action: ExecuteContract{
				ContractAddress:     testAddress,
				Payload:             nil,
				Keys:                nil,
				ComputeUnitsToSpend: 123456,
				FunctionName:        "",
			},
		},
		{
			name: "Keys",
			action: ExecuteContract{
				ContractAddress: testAddress,
				Payload:         nil,
				Keys: map[string]state.Permissions{
					string([]byte{0x01, 0x02, 0x03, 0x04}): state.Read,
				},
				ComputeUnitsToSpend: 0,
				FunctionName:        "",
			},
		},
		{
			name: "All fields filled",
			action: ExecuteContract{
				ContractAddress: testAddress,
				Payload:         []byte("payload"),
				Keys: map[string]state.Permissions{
					string([]byte{0x01, 0x02, 0x03, 0x04}): state.Read | state.Write,
					string([]byte{0x05, 0x06, 0x07, 0x08}): state.Write,
					string([]byte{0x09, 0x0a, 0x0b, 0x0c}): state.Allocate,
					string([]byte{0x0d, 0x0e, 0x0f, 0x10}): state.All,
				},
				ComputeUnitsToSpend: 999888777,
				FunctionName:        "functionName2",
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
			if tt.errorExpected {
				require.Error(t, err)
				return //all good, catch error and exit this case
			} else {
				require.NoError(t, err)
			}

			require.NotNil(t, unmarshalledAction)
			unmarshalledEC := unmarshalledAction.(*ExecuteContract)

			require.Equal(t, tt.action.ContractAddress, unmarshalledEC.ContractAddress, "ContractAddress mismatch")
			require.Equal(t, tt.action.Payload, unmarshalledEC.Payload, "Payload mismatch")
			require.Equal(t, len(tt.action.Keys), len(unmarshalledEC.Keys), "Keys length mismatch")
			require.Equal(t, tt.action.ComputeUnitsToSpend, unmarshalledEC.ComputeUnitsToSpend, "ComputeUnitsToSpend mismatch")
			require.Equal(t, tt.action.FunctionName, unmarshalledEC.FunctionName, "FunctionName mismatch")

			for k, v := range tt.action.Keys {
				require.Equal(t, v, unmarshalledEC.Keys[k], "Permissions mismatch for key %v", k)
			}
		})
	}
}
