

interface JavyBuiltins {
    IO: {
        readSync(fd: number, buffer: Uint8Array): number;
        writeSync(fd: number, buffer: Uint8Array): number;
    };
}

declare global {
    const Javy: JavyBuiltins;
    function callback(payload: Uint8Array): Uint8Array;
}


export type GetBytesFunc = (slot: Uint8Array, chunks: number) => Uint8Array;
export type SetBytesFunc = (slot: Uint8Array, chunks: number, value: Uint8Array) => void;

export type ExecuteContractFunc = (payload: Uint8Array, actor: Uint8Array, getBytes: GetBytesFunc, setBytes: SetBytesFunc) => Uint8Array;
