package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/ava-labs/hypersdk/examples/typescriptvm/cmd/wasmsdk/wasmsdk"
)

func generateDigest() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) != 1 {
			return js.ValueOf("Invalid number of arguments passed")
		}
		inputJSON := args[0].String()

		var unmarshaled wasmsdk.Base
		err := json.Unmarshal([]byte(inputJSON), &unmarshaled)
		if err != nil {
			fmt.Println("Error unmarshaling from JSON:", err)
			return js.ValueOf(fmt.Sprintf("Error unmarshaling from JSON: %s", err.Error()))
		}

		digest, err := wasmsdk.Digest(&unmarshaled)
		if err != nil {
			fmt.Println("Error generating digest:", err)
			return js.ValueOf(fmt.Sprintf("Error generating digest: %s", err.Error()))
		}

		return js.ValueOf(fmt.Sprintf("%x", digest))
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("generateDigest", generateDigest())
	select {}
}
