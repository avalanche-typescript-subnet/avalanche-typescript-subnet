package runtime_test

//go:generate npx ../../../runtime/js_sdk assets/simple_counter.ts
//go:generate cp assets/simple_counter.wasm ../actions/assets/simple_counter_copy.wasm

import (
	_ "embed"

	"github.com/ava-labs/hypersdk/codec"
)

//go:embed assets/simple_counter.wasm
var testWasmBytes []byte

func createActorAddress(actorNumber uint) []byte {
	actor := codec.Address{byte(actorNumber), byte(actorNumber), byte(actorNumber)}
	actorBytes := make([]byte, len(actor))
	copy(actorBytes, actor[:])
	return actorBytes
}

// func (d *DummyStateProvider) Print() {
// 	fmt.Printf("State now has %d entries\n", len(d.State))
// 	for k, v := range d.State {
// 		fmt.Printf("%x: %x\n", k, v)
// 	}
// 	fmt.Println()
// }
