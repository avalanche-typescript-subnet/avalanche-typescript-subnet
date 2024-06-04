package runtime_test

//go:generate npx ../../../runtime/js_sdk test_assets/simpleCounter.ts

import (
	_ "embed"
	"fmt"

	"github.com/ava-labs/hypersdk/codec"
)

const CONTRACT_ACTION_READ = 0
const CONTRACT_ACTION_INCREMENT = 1
const CONTRACT_ACTION_LOAD_CPU = 2
const CONTRACT_ACTION_WRITE_MANY_SLOTS = 3
const CONTRACT_ACTION_READ_MANY_SLOTS = 4
const CONTRACT_ACTION_ECHO = 5

//go:embed test_assets/simpleCounter.wasm
var testWasmBytes []byte

func createActorAddress(actorNumber uint) []byte {
	actor := codec.Address{byte(actorNumber), byte(actorNumber), byte(actorNumber)}
	actorBytes := make([]byte, len(actor))
	copy(actorBytes, actor[:])
	return actorBytes
}

type DummyStateProvider struct {
	State map[[4]byte][]byte
}

func (d *DummyStateProvider) StateProvider(key [4]byte) ([]byte, error) {
	value, exists := d.State[key]
	if !exists {
		return []byte{}, nil
	}
	return value, nil
}

func (d *DummyStateProvider) SetState(key [4]byte, value []byte) {
	d.State[key] = value
	d.Print()
}

func (d *DummyStateProvider) Print() {
	fmt.Printf("State now has %d entries\n", len(d.State))
	for k, v := range d.State {
		fmt.Printf("%x: %x\n", k, v)
	}
	fmt.Println()
}
