package runtime

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v21"
)

type JavyExecResult struct {
	FuelConsumed uint64
	TimeTaken    time.Duration
	Result       ResultJSON
	DebugLog     []byte
}

type JavyExec struct {
	executeMutexes map[uint64]*sync.Mutex
	storesCache    map[uint64]*struct {
		store    *wasmtime.Store
		mainFunc *wasmtime.Func
	}
}

func NewJavyExec() *JavyExec {
	return &JavyExec{
		executeMutexes: map[uint64]*sync.Mutex{},
		storesCache: map[uint64]*struct {
			store    *wasmtime.Store
			mainFunc *wasmtime.Func
		}{},
	}
}

func (exec *JavyExec) Execute(params JavyExecParams) (*JavyExecResult, error) {
	state := make(map[string][]byte)

	for i := 0; i < 100; i++ { //curcuit breaker
		res, err := exec.executeOnState(params, state)
		if err != nil {
			return nil, err
		}

		if strings.Contains(res.Result.Error, "NO_VALUE_AT_ADDRESS") {
			addrHex := strings.Split(res.Result.Error, "\"")[1]

			addressBytes, err := hex.DecodeString(strings.TrimPrefix(addrHex, "0x"))
			if err != nil {
				return nil, fmt.Errorf("error decoding address %s: %v", addrHex, err)
			}

			newVal, err := params.StateProvider(string(addressBytes))
			if err != nil {
				return nil, fmt.Errorf("error retrieving state for address %x: %v", addressBytes, err)
			}
			state[string(addressBytes)] = newVal

			fmt.Printf("state %+v", state)
		} else {
			return res, nil
		}
	}

	return nil, fmt.Errorf("execution failed after 100 attempts")
}

func (exec *JavyExec) executeOnState(params JavyExecParams, state map[string][]byte) (*JavyExecResult, error) {
	store, mainFunc, err := exec.createStore(params.Bytecode)
	if err != nil {
		return nil, err
	}

	store.Limiter(params.MaxMemory, 629, 2, 1, 1)

	if !EXECUTE_CACHE_ENABLED {
		defer store.Engine.Close()
		defer store.Close()
	}

	callDataJson, err := json.Marshal(JSPayload{
		CurrentState: &state,
		Payload:      params.Payload,
		FunctionName: params.FunctionName,
		Actor:        params.Actor,
	})

	if err != nil {
		return nil, fmt.Errorf("marshalling call data: %v", err)
	}

	//FIXME: use syscall.Mkfifo instead of torturing disk with temp files
	stdoutFile, err := os.CreateTemp("", "stdout*")
	if err != nil {
		return nil, fmt.Errorf("creating stdout file: %v", err)
	}
	defer os.Remove(stdoutFile.Name())

	stderrFile, err := os.CreateTemp("", "stderr*")
	if err != nil {
		return nil, fmt.Errorf("creating stderr file: %v", err)
	}
	defer os.Remove(stderrFile.Name())

	stdinFile, err := os.CreateTemp("", "stdin*")
	if err != nil {
		return nil, fmt.Errorf("creating stdin file: %v", err)
	}
	defer os.Remove(stdinFile.Name())

	_, err = stdinFile.Write(callDataJson)
	if err != nil {
		return nil, fmt.Errorf("writing to stdin file: %v", err)
	}

	err = store.SetFuel(params.MaxFuel)
	if err != nil {
		return nil, fmt.Errorf("setting fuel: %v", err)
	}

	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutFile.Name())
	wasiConfig.SetStderrFile(stderrFile.Name())
	wasiConfig.SetStdinFile(stdinFile.Name())
	store.SetWasi(wasiConfig)

	startTime := time.Now()

	finished := false

	timeoutErrCh := make(chan error, 1)
	go func() {
		time.Sleep(params.MaxTime)
		if !finished {
			fmt.Printf("Execution timed out\n")
			store.Engine.IncrementEpoch()
			timeoutErrCh <- fmt.Errorf("execution timed out")
		}
	}()

	_, err = mainFunc.Call(store)
	finished = true
	if err != nil {
		return nil, fmt.Errorf("calling user code main function: %v", err)
	}
	execTime := time.Since(startTime)

	select {
	case err := <-timeoutErrCh:
		if err != nil {
			return nil, err
		}
	default:
	}

	fuelAfter, err := store.GetFuel()
	if err != nil {
		return nil, fmt.Errorf("getting fuel after execution: %v", err)
	}
	consumedFuel := params.MaxFuel - fuelAfter

	stdoutBytes, err := os.ReadFile(stdoutFile.Name())
	if err != nil {
		return nil, fmt.Errorf("reading stdout file: %v", err)
	}

	fmt.Printf("\n>>>>DEBUG stdout>>>\n> %s\n<<<<DEBUG stdout<<<<\n\n", strings.ReplaceAll(string(stdoutBytes), "\n", "\n> "))

	stderrBytes, err := os.ReadFile(stderrFile.Name())
	if err != nil {
		return nil, fmt.Errorf("reading stderr file: %v", err)
	}

	fmt.Printf("\n>>>>DEBUG stderr>>>\n> %s\n<<<<DEBUG stderr<<<<\n\n", strings.ReplaceAll(string(stderrBytes), "\n", "\n> "))

	var stdoutResult ResultJSON
	err = json.Unmarshal(stdoutBytes, &stdoutResult)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling stdout: %v", err)
	}

	return &JavyExecResult{
		FuelConsumed: consumedFuel,
		TimeTaken:    execTime,
		Result:       stdoutResult,
		DebugLog:     stderrBytes,
	}, nil
}
