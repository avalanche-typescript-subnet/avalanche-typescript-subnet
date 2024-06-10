import { Base64ToHexString, Base64ToUint8Array, HexStringToBase64, Uint8ArrayToBase64 } from "./encoders";
import { readStdin, writeStdOut } from "./javy_io";
import { ExecuteContractFunc } from "./types";

const keyHexAddress = (slot: number, chunks: number) => {
    //check slot and chunks are valid
    if (slot < 0 || slot > 65535) {
        throw new Error(`Slot must be a value between 0 and 65535.`);
    }
    if (chunks < 1 || chunks > 65535) {//uint16 check for size
        throw new Error("Size must be a value between 1 and 65535.");
    }

    return `0x${slot.toString(16).padStart(4, '0')}${chunks.toString(16).padStart(4, '0')}`
}

console.warn = console.log//FIXME: monkey-patching

class TracebleDB {
    private _state: Record<string, Uint8Array>;
    private _reads: Set<string>;
    private _writes: Set<string>;

    constructor(state: Record<string, Uint8Array>) {
        this._state = state;
        this._reads = new Set();
        this._writes = new Set();
    }

    getValue(slot: number, chunks: number): Uint8Array {
        const address = keyHexAddress(slot, chunks);
        console.log(`Reading ${address}`)

        this._reads.add(address);

        if (!this._state[address]) {
            throw new Error(`NO_VALUE_AT_ADDRESS="${address}"`)
        }

        return this._state[address];
    }

    setValue(slot: number, chunks: number, value: Uint8Array): void {
        const address = keyHexAddress(slot, chunks);
        console.log(`Writing ${address}`)
        //check size
        if (value.length > 64 * chunks) {
            throw new Error(`Value is larger than the size of the key (max size: ${64 * chunks} in ${chunks} 64-byte chunks, value: ${value.length})`);
        }

        if (this._state[address]
            && (this._state[address].length === value.length)
            && this._state[address].every((byte, i) => byte === value[i])
        ) {
            console.warn(`Skipping write of ${address} because it is already set to the same value.`);
            return;
        }

        this._writes.add(address);

        this._state[address] = value
    }

    getReadAdresses(): string[] {
        return [...this._reads];
    }

    getWriteAddresses(): Record<string, Uint8Array> {
        const result: Record<string, Uint8Array> = {};
        this._writes.forEach(address => {
            result[address] = this._state[address];
        });
        return result;
    }
}

let isFunctionRegistered = false;
export function registerRawFunction(func: ExecuteContractFunc) {
    if (isFunctionRegistered) {
        throw new Error("registerFunction can only be called once.");
    }
    isFunctionRegistered = true;

    try {

        const stdin = readStdin();
        const argsJSON = JSON.parse(stdin) as {
            currentState: Record<string, string>,
            payload: string,
            actor: string,
        }

        const stateBytes: Record<string, Uint8Array> = {};
        for (const [key, value] of Object.entries(argsJSON.currentState)) {
            stateBytes[Base64ToHexString(key)] = Uint8Array.from(Base64ToUint8Array(value));
        }

        const db = new TracebleDB(stateBytes);
        const actor = Base64ToUint8Array(argsJSON.actor);
        const payload = Base64ToUint8Array(argsJSON.payload);

        const result = func(payload, actor, db.getValue.bind(db), db.setValue.bind(db))

        const updatedKeys: Record<string, string> = {};
        for (const [key, value] of Object.entries(db.getWriteAddresses())) {
            updatedKeys[HexStringToBase64(key)] = Uint8ArrayToBase64(value);
        }

        writeStdOut(JSON.stringify({
            success: true,
            result: Uint8ArrayToBase64(result),
            readKeys: db.getReadAdresses().map(key => HexStringToBase64(key)),
            updatedKeys,
        }))
    } catch (e) {
        writeStdOut(JSON.stringify({
            success: false,
            error: String(e) + "\n" + (e as Error)?.stack,
        }))
    }
}

export * as encoders from './encoders';
export * as io from './javy_io';
export * as types from './types.d';
