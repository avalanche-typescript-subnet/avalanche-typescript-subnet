package runtime

import (
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
	callback CallbackFunc
}

func NewJavyExec() *JavyExec {
	return &JavyExec{
		executeMutexes: map[uint64]*sync.Mutex{},
		storesCache: map[uint64]*struct {
			store    *wasmtime.Store
			mainFunc *wasmtime.Func
		}{},
		callback: EmptyCallbackFunc,
	}
}

func (exec *JavyExec) Execute(params JavyExecParams) (*JavyExecResult, error) {
	state := make(map[string][]byte)

	res, err := exec.executeOnState(params, state)
	if err != nil {
		return nil, err
	}

	return res, nil
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
