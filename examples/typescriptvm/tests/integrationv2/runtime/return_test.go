package runtime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
)

func TestReturn(t *testing.T) {
	exec := runtime.NewJavyExec()
	stateProvider := &DummyStateProvider{State: map[[4]byte][]byte{{0, 0, 0, 0xA}: {0x7b, 0x7d}}}

	actor1Bytes := createActorAddress(1)
	actor2Bytes := createActorAddress(2)
	actor3Bytes := createActorAddress(3)

	params := runtime.JavyExecParams{
		MaxFuel:       10 * 1000 * 1000,
		MaxTime:       time.Millisecond * 100,
		MaxMemory:     1024 * 1024 * 100,
		Bytecode:      &testWasmBytes,
		StateProvider: stateProvider.StateProvider,
		Payload:       append([]byte{1}, 7),
	}

	//execute 2 times for actor 1
	params.Actor = actor1Bytes
	for i := 0; i < 2; i++ {
		res, err := exec.Execute(params)
		if err != nil {
			t.Fatal(err)
		}
		for updatedKey, updatedVal := range res.Result.UpdatedKeys {
			stateProvider.SetState(updatedKey, updatedVal)
		}
	}

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
	params.Payload = []byte{0}
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
	params.Payload = []byte{0}
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
	params.Payload = []byte{0}
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
