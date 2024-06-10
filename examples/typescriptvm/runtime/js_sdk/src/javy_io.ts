const enum STDIO {
    Stdin,
    Stdout,
    Stderr,
}

export function readStdin(): string {
    let buffer = new Uint8Array(1024);
    let bytesUsed = 0;
    while (true) {
        const bytesRead = Javy.IO.readSync(STDIO.Stdin, buffer.subarray(bytesUsed));
        // A negative number of bytes read indicates an error.
        if (bytesRead < 0) {
            // FIXME: Figure out the specific error that occured.
            throw Error("Error while reading from file descriptor");
        }
        // 0 bytes read means we have reached EOF.
        if (bytesRead === 0) {
            const endBuffer = buffer.subarray(0, bytesUsed + bytesRead);
            return new TextDecoder().decode(endBuffer);
        }

        bytesUsed += bytesRead;
        // If we have filled the buffer, but have not reached EOF yet,
        // double the buffers capacity and continue.
        if (bytesUsed === buffer.length) {
            const nextBuffer = new Uint8Array(buffer.length * 2);
            nextBuffer.set(buffer);
            buffer = nextBuffer;
        }
    }
}

export function writeStdOut(input: string) {
    const encoder = new TextEncoder();
    const buffer = encoder.encode(input);
    writeFileSync(STDIO.Stdout, buffer);
}

export function writeStdErr(input: string) {
    const encoder = new TextEncoder();
    const buffer = encoder.encode(input);
    writeFileSync(STDIO.Stderr, buffer);
}

function writeFileSync(fd: number, buffer: Uint8Array) {
    while (buffer.length > 0) {
        // Try to write the entire buffer.
        const bytesWritten = Javy.IO.writeSync(fd, buffer);
        // A negative number of bytes written indicates an error.
        if (bytesWritten < 0) {
            throw Error("Error while writing to file descriptor");
        }
        // 0 bytes means that the destination cannot accept additional bytes.
        if (bytesWritten === 0) {
            throw Error("Could not write all contents in buffer to file descriptor");
        }
        // Otherwise cut off the bytes from the buffer that
        // were successfully written.
        buffer = buffer.subarray(bytesWritten);
    }
}

