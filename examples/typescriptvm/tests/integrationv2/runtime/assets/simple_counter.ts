import { registerFunc, encoders, GetBytesFunc, SetBytesFunc, execute } from "../../../../runtime/js_sdk";

const decoder = new TextDecoder();
const encoder = new TextEncoder();

registerFunc("read", (_payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, _setBytes: SetBytesFunc) => {
    const actorString = encoders.Uint8ArrayToBase64(actor);

    const stringAtSlotZero = decoder.decode(
        getBytes(new Uint8Array([0x0]), 10)
    )
    const dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

    const toReturn = dataAtSlotZero[actorString] || 0

    return encoder.encode(JSON.stringify(toReturn));
})

registerFunc("increment", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const actorString = encoders.Uint8ArrayToBase64(actor);
    const value = payload[0];


    const stringAtSlotZero = decoder.decode(
        getBytes(new Uint8Array([0x0]), 10)
    )
    const dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

    dataAtSlotZero[actorString] = (dataAtSlotZero[actorString] || 0) + value;

    setBytes(new Uint8Array([0x0]), 10, encoder.encode(JSON.stringify(dataAtSlotZero)));

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
