
import { registerFunc, encoders, GetBytesFunc, SetBytesFunc, execute } from "../../../../runtime/js_sdk";

const encoder = new TextEncoder();

registerFunc("test_callback", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    let res = callback(payload);
    return encoder.encode(JSON.stringify(res));
})

execute();