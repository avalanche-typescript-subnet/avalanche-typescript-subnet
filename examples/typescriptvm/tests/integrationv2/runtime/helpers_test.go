package runtime_test

//--go:generate npx ../../../runtime/js_sdk assets/simple_counter.ts
//go:generate ../../../runtime/js_wcb_sdk/build.sh assets/simple_counter.ts
//go:generate cp assets/simple_counter.wasm ../actions/assets/simple_counter_copy.wasm

import (
	_ "embed"
	"fmt"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/runtime"
)

//go:embed assets/simple_counter.wasm
var testWasmBytes []byte

func createActorAddress(actorNumber uint) []byte {
	actor := codec.Address{byte(actorNumber), byte(actorNumber), byte(actorNumber)}
	actorBytes := make([]byte, len(actor))
	copy(actorBytes, actor[:])
	return actorBytes
}

type DummyStateProvider struct {
	State map[runtime.KeyPostfix][]byte
}

func (d *DummyStateProvider) StateProvider(key runtime.KeyPostfix) ([]byte, error) {
	value, exists := d.State[key]
	if !exists {
		return []byte{}, nil
	}
	return value, nil
}

func (d *DummyStateProvider) SetState(key runtime.KeyPostfix, value []byte) {
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
