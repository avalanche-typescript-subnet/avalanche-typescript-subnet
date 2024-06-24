package runtime_test

//go:generate ../../../runtime/js_wcb_sdk/build.sh compile assets/simple_counter.ts
//go:generate ../../../runtime/js_wcb_sdk/build.sh emit-provider ../../../runtime/javy_provider.wasm

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
