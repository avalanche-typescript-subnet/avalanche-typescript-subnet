package runtime_test

import (
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
	Payload:       []byte{CONTRACT_ACTION_READ},
	Actor:         []byte{},
	StateProvider: (&DummyStateProvider{}).StateProvider,
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
		{"0x10, 0,  error", 0, []byte{CONTRACT_ACTION_LOAD_CPU, 0x10}, true},
		{"0x10, 125M, no error", 125 * MIL, []byte{CONTRACT_ACTION_LOAD_CPU, 0x10}, false},
		{"0x11, 125M, error", 125 * MIL, []byte{CONTRACT_ACTION_LOAD_CPU, 0x11}, true},
		{"0x11, 130M, no error", 130 * MIL, []byte{CONTRACT_ACTION_LOAD_CPU, 0x11}, false},
		{"0x12, 130M, error", 130 * MIL, []byte{CONTRACT_ACTION_LOAD_CPU, 0x12}, true},
		{"0x12, 150M, no error", 150 * MIL, []byte{CONTRACT_ACTION_LOAD_CPU, 0x12}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params.MaxFuel = tc.maxFuel
			params.Payload = tc.payload
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
	params.Payload = []byte{CONTRACT_ACTION_WRITE_MANY_SLOTS, 0x10}

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
		name          string
		maxMemory     int64
		payload       []byte
		expectSuccess bool
	}{
		{"20 pages, 0x17 - no error expected", 20 * MEMORY_PAGE_64KB, []byte{0x02, 0x17}, true},
		{"20 pages, 0x18 - error expected", 20 * MEMORY_PAGE_64KB, []byte{0x02, 0x18}, false},
		{"21 pages, 0x18 - no error expected", 21 * MEMORY_PAGE_64KB, []byte{0x02, 0x18}, true},
		{"21 pages, 0x19 - no error expected", 21 * MEMORY_PAGE_64KB, []byte{0x02, 0x19}, true},
		//...
		{"21 pages, 0x27 - error expected", 21 * MEMORY_PAGE_64KB, []byte{0x02, 0x27}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
