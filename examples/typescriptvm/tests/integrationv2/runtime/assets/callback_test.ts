
import { registerFunc, encoders, GetBytesFunc, SetBytesFunc, execute } from "../../../../runtime/js_sdk";
import { readStdin,writeStdOut } from "../../../../runtime/js_sdk/src/javy_io";
import { Base64ToUint8Array, Uint8ArrayToBase64, } from "../../../../runtime/js_sdk/src/encoders";

const encoder = new TextEncoder();

// registerFunc("test_callback", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
//     let res = callback(payload);
//     return encoder.encode(JSON.stringify(res));
// })

// execute();

const stdin = readStdin();
        const argsJSON = JSON.parse(stdin) as {
            currentState: Record<string, string>,
            payload: string,
            actor: string,
            functionName: string,
        }

const payload = Base64ToUint8Array(argsJSON.payload);

let res = callback(payload)

for (let i = 0; i < res.length; i++) {
    res[i] = res[i] + 1
}

if (payload[0] == 0) {
    writeStdOut(JSON.stringify({
        success: true,
        result: [1],
        readKeys: null,
        updatedKeys: null,
        error: "",
    }))    
} else {
    writeStdOut(JSON.stringify({
        success: true,
        result: Array.from(res),
        readKeys: null,
        updatedKeys: null,
        error: "",
    }))    
}

