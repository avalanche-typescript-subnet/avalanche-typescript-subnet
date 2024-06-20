package runtime_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"

	"github.com/stretchr/testify/require"
)

var DEFAULT_PARAMS_LIMITS = runtime.JavyExecParams{
	MaxFuel:       10 * 1000 * 1000,
	MaxTime:       time.Millisecond * 200,
	MaxMemory:     1024 * 1024 * 100,
	Bytecode:      &testWasmBytes,
	Payload:       []byte{},
	FunctionName:  "loadCPU",
	Actor:         []byte{},
	StateProvider: runtime.NewDummyStateProvider().StateProvider,
}

func TestMaxFuel(t *testing.T) {
	exec := runtime.NewJavyExec()

	const MIL = 1 * 1000 * 1000
	params := DEFAULT_PARAMS_LIMITS

	testCases := []struct {
		name        string
		maxFuel     uint64
		payload     []byte
		expectError bool
	}{
		{"0x10, 0,  error", 0, []byte{0x10}, true},
		{"0x10, 125M, no error", 125 * MIL, []byte{0x10}, false},
		{"0x11, 125M, error", 125 * MIL, []byte{0x11}, true},
		{"0x11, 130M, no error", 130 * MIL, []byte{0x11}, false},
		{"0x12, 130M, error", 130 * MIL, []byte{0x12}, true},
		{"0x12, 150M, no error", 150 * MIL, []byte{0x12}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params.MaxFuel = tc.maxFuel
			params.Payload = tc.payload
			params.FunctionName = "loadCPU"
			res, err := exec.Execute(params)

			if err != nil && !strings.Contains(err.Error(), "all fuel consumed") {
				t.Errorf("Expected an all fuel consumed error but got %v", err)
			}

			if (err != nil) != tc.expectError {
				if tc.expectError {
					t.Errorf("Expected an error but got none; %vM consumed", res.FuelConsumed/MIL)
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}

}

func TestMaxTime(t *testing.T) {
	exec := runtime.NewJavyExec()

	params := DEFAULT_PARAMS_LIMITS
	params.Payload = []byte{0x10}
	params.FunctionName = "writeManySlots"

	_, err := exec.Execute(params)
	if err != nil {
		t.Error(err)
		return
	}

	params.MaxTime = time.Nanosecond

	_, err = exec.Execute(params)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestMaxMemory(t *testing.T) {
	const MEMORY_PAGE_64KB = 1024 * 64

	exec := runtime.NewJavyExec()

	params := DEFAULT_PARAMS_LIMITS
	params.MaxFuel = 1000 * 1000 * 1000

	testCases := []struct {
		maxMemory     int64
		payload       []byte
		expectSuccess bool
	}{
		{22 * MEMORY_PAGE_64KB, []byte{0x19}, true},
		{22 * MEMORY_PAGE_64KB, []byte{0x20}, false},
		{23 * MEMORY_PAGE_64KB, []byte{0x20}, true},
		{23 * MEMORY_PAGE_64KB, []byte{0x26}, true},
		{23 * MEMORY_PAGE_64KB, []byte{0x27}, false},
		{24 * MEMORY_PAGE_64KB, []byte{0x27}, true},
		//...
		{25 * MEMORY_PAGE_64KB, []byte{0x29}, true},
		{25 * MEMORY_PAGE_64KB, []byte{0x30}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%d pages, payload=0x%x, expectedSuccess=%v", tc.maxMemory/MEMORY_PAGE_64KB, tc.payload, tc.expectSuccess), func(t *testing.T) {
			params.MaxMemory = tc.maxMemory
			params.Payload = tc.payload
			res, err := exec.Execute(params)

			require.Nil(t, err)

			require.Equal(t, res.Result.Success, tc.expectSuccess)

			if !res.Result.Success {
				require.Contains(t, res.Result.Error, "out of memory")
			}
		})
	}
}
