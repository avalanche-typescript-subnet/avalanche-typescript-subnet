import { registerFunc, GetBytesFunc, SetBytesFunc, execute } from "../../../../runtime/js_sdk/src/runtime";
import { BigintToUint8Array, Uint8ArrayToBigint } from "../../../../runtime/js_sdk/src/runtime/encoders";

const decoder = new TextDecoder();
const encoder = new TextEncoder();

registerFunc("read", (_payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, _setBytes: SetBytesFunc) => {
    const balanceBytes = getBytes(new Uint8Array(actor), 10)
    const balance = Uint8ArrayToBigint(balanceBytes)

    return encoder.encode(balance.toString());
})

registerFunc("increment", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const value = payload[0];

    const balanceBytes = getBytes(new Uint8Array(actor), 10)
    const balance = Uint8ArrayToBigint(balanceBytes)
    const newBalance = balance + BigInt(value)
    setBytes(new Uint8Array(actor), 10, BigintToUint8Array(newBalance))

    return new Uint8Array();
})

registerFunc("echo", (payload: Uint8Array, _actor: Uint8Array, _getBytes: GetBytesFunc, _setBytes: SetBytesFunc) => {
    const value = payload[0];

    return encoder.encode(JSON.stringify(value));
})


registerFunc("writeManySlots", (payload: Uint8Array, _actor: Uint8Array, _getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const value = payload[0];

    for (let i = 1; i <= value; i++) {
        setBytes(new Uint8Array([i]), 6, Uint8Array.from([i, i, i]));
    }

    return new Uint8Array();
})


registerFunc("readManySlots", (payload: Uint8Array, _actor: Uint8Array, getBytes: GetBytesFunc, _setBytes: SetBytesFunc) => {
    const value = payload[0];

    let result = new Uint8Array();
    for (let i = 0; i < value; i++) {
        const fromSlot = getBytes(new Uint8Array([i + 1]), 6);
        let tempResult = new Uint8Array(result.length + fromSlot.length);
        tempResult.set(result);
        tempResult.set(fromSlot, result.length);
        result = tempResult;
    }
    return result;
})

registerFunc("loadCPU", (payload: Uint8Array, _actor: Uint8Array, _getBytes: GetBytesFunc, _setBytes: SetBytesFunc) => {
    const value = payload[0];

    const memoryBalloon = new Array(value * value * 1024);
    for (let i = 0; i < value; i++) {
        for (let j = 0; j < value; j++) {
            for (let k = 0; k < value; k++) {
                memoryBalloon[i * j * k] = i * j * k;
            }
        }
    }
    const summ = memoryBalloon.reduce((acc, curr) => acc + curr, 0);
    return encoder.encode(JSON.stringify(summ));
})



execute();
