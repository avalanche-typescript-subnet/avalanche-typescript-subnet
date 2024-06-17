import { registerFunc, encoders, GetBytesFunc, SetBytesFunc, execute } from "../../../../runtime/js_sdk";

const decoder = new TextDecoder();
const encoder = new TextEncoder();

const CONTRACT_ACTION_READ = 0;
const CONTRACT_ACTION_INCREMENT = 1;
const CONTRACT_ACTION_LOAD_CPU = 2;
const CONTRACT_ACTION_WRITE_MANY_SLOTS = 3;
const CONTRACT_ACTION_READ_MANY_SLOTS = 4;
const CONTRACT_ACTION_ECHO = 5;

registerFunc("read", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const actorString = encoders.Uint8ArrayToBase64(actor);

    const stringAtSlotZero = decoder.decode(
        getBytes(0, 10)
    )
    const dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

    const toReturn = dataAtSlotZero[actorString] || 0

    return encoder.encode(JSON.stringify(toReturn));
})

registerFunc("increment", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const actorString = encoders.Uint8ArrayToBase64(actor);
    const value = payload[0];


    const stringAtSlotZero = decoder.decode(
        getBytes(0, 10)
    )
    const dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

    dataAtSlotZero[actorString] = (dataAtSlotZero[actorString] || 0) + value;

    setBytes(0, 10, encoder.encode(JSON.stringify(dataAtSlotZero)));

    return new Uint8Array();
})

registerFunc("echo", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const value = payload[0];

    return encoder.encode(JSON.stringify(value));
})


registerFunc("writeManySlots", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const value = payload[0];

    for (let i = 1; i <= value; i++) {
        setBytes(i, 6, Uint8Array.from([i, i, i]));
    }

    return new Uint8Array();
})


registerFunc("readManySlots", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
    const value = payload[0];

    let result = new Uint8Array();
    for (let i = 0; i < value; i++) {
        const fromSlot = getBytes(i + 1, 6);
        let tempResult = new Uint8Array(result.length + fromSlot.length);
        tempResult.set(result);
        tempResult.set(fromSlot, result.length);
        result = tempResult;
    }
    return result;
})

registerFunc("loadCPU", (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => {
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
