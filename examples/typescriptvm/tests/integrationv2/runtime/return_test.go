package runtime_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
	"github.com/stretchr/testify/require"
)

func TestReturn(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateProvider := &DummyStateProvider{State: map[[4]byte][]byte{}}

	actor1Bytes := createActorAddress(1)
	actor2Bytes := createActorAddress(2)
	actor3Bytes := createActorAddress(3)

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateProvider.StateProvider,
		Payload:       []byte{CONTRACT_ACTION_INCREMENT, 7},
	}

	//execute 2 times for actor 1
	params.Actor = actor1Bytes
	for i := 0; i < 2; i++ {
		res, err := exec.Execute(params)
		if err != nil {
			t.Fatal(err)
		}

		require.Equal(t, true, res.Result.Success)

		fmt.Printf("UpdatedKeys: %v\n", res.Result.UpdatedKeys)
		for updatedKey, updatedVal := range res.Result.UpdatedKeys {
			stateProvider.SetState(updatedKey, updatedVal)
		}
	}
	stateProvider.Print()

	//execute 1 time for actor 2
	params.Actor = actor2Bytes
	res, err := exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateProvider.SetState(updatedKey, updatedVal)
	}

	//get result for actor 1
	params.Payload = []byte{CONTRACT_ACTION_READ}
	params.Actor = actor1Bytes

	res, err = exec.Execute(params)
	if err != nil {
		t.Fatal(err)
	}
	for updatedKey, updatedVal := range res.Result.UpdatedKeys {
		stateProvider.SetState(updatedKey, updatedVal)
	}

	//check result for actor 1
	var actor1FinalBalance uint32
	if err := json.Unmarshal(res.Result.Result, &actor1FinalBalance); err != nil {
		t.Fatal(err)
	}
	if actor1FinalBalance != 7*2 { //2 increments, 7 each
		t.Fatalf("Expected balance %d, got %d", 7*2, actor1FinalBalance)
	}

	//get result for actor 2
	params.Payload = []byte{CONTRACT_ACTION_READ}
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
	params.Payload = []byte{CONTRACT_ACTION_READ}
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
	params.Payload = []byte{CONTRACT_ACTION_ECHO, 124}
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
