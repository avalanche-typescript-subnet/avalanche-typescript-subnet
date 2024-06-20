package runtime_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/stretchr/testify/require"
)

func TestReturn(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateProvider := runtime.NewDummyStateProvider()

	actor1Bytes := createActorAddress(1)
	actor2Bytes := createActorAddress(2)
	actor3Bytes := createActorAddress(3)

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateProvider.StateProvider,
		Payload:       []byte{7},
		FunctionName:  "increment",
	}

	var callback runtime.CallbackFunc = func(address []byte) ([]byte, error) {
		addrHex := string(address)

		addressBytes, err := hex.DecodeString(strings.TrimPrefix(addrHex, "0x"))
		if err != nil {
			return nil, fmt.Errorf("error decoding address %s: %v", addrHex, err)
		}

		fmt.Println("addressBytes", string(addressBytes))
		newVal, err := stateProvider.StateProvider(string(addressBytes))
		if err != nil {
			return nil, fmt.Errorf("error retrieving state for address %x: %v", addressBytes, err)
		}
		return newVal, nil
	}

	exec.SetCallbackFunc(callback)

	//execute 2 times for actor 1
	params.Actor = actor1Bytes
	for i := 0; i < 2; i++ {
		res, err := exec.Execute(params)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, true, res.Result.Success)

		fmt.Printf("UpdatedKeys: %v\n", res.Result.UpdatedKeys)
		stateProvider.Update(res.Result.UpdatedKeys)
	}

	//execute 1 time for actor 2
	params.Actor = actor2Bytes
	res, err := exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	stateProvider.Update(res.Result.UpdatedKeys)

	//get result for actor 1
	params.Payload = []byte{}
	params.FunctionName = "read"
	params.Actor = actor1Bytes

	res, err = exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	stateProvider.Update(res.Result.UpdatedKeys)

	//check result for actor 1
	var actor1FinalBalance uint32
	if err := json.Unmarshal(res.Result.Result, &actor1FinalBalance); err != nil {
		t.Fatal(err)
	}
	if actor1FinalBalance != 7*2 { //2 increments, 7 each
		t.Fatalf("Expected balance %d, got %d", 7*2, actor1FinalBalance)
	}

	//get result for actor 2
	params.Payload = []byte{}
	params.FunctionName = "read"
	params.Actor = actor2Bytes
	res, err = exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}

	//check result for actor 2
	var actor2FinalBalance uint32
	if err := json.Unmarshal(res.Result.Result, &actor2FinalBalance); err != nil {
		t.Fatal(err)
	}

	if actor2FinalBalance != 7 {
		t.Fatalf("Expected balance %d, got %d", 7, actor2FinalBalance)
	}

	//check zero balance for actor 3
	params.Payload = []byte{}
	params.FunctionName = "read"
	params.Actor = actor3Bytes
	res, err = exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}

	var actor3FinalBalance uint32
	if err := json.Unmarshal(res.Result.Result, &actor3FinalBalance); err != nil {
		t.Fatal(err)
	}

	if actor3FinalBalance != 0 {
		t.Fatalf("Expected balance %d, got %d", 0, actor3FinalBalance)
	}
}

func TestEcho(t *testing.T) {
	exec := runtime.NewJavyExec()
	params := DEFAULT_PARAMS_LIMITS
	params.Payload = []byte{124}
	params.FunctionName = "echo"
	res, err := exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}

	var result uint32
	if err := json.Unmarshal(res.Result.Result, &result); err != nil {
		t.Fatal(err)
	}
	if result != 124 {
		t.Fatalf("Expected result %d, got %d", 124, result)
	}
}
