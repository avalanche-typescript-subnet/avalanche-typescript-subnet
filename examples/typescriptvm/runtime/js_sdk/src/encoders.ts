const BASE64_CHARS = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';

export function Uint8ArrayToBase64(uint8Array: Uint8Array): string {
    let bitString = '';
    // Convert each byte to its 8-bit binary representation
    for (let i = 0; i < uint8Array.length; i++) {
        bitString += uint8Array[i].toString(2).padStart(8, '0');
    }

    let base64 = '';
    // Process each 6-bit segment
    for (let i = 0; i < bitString.length; i += 6) {
        const bits = bitString.substring(i, i + 6);
        // Right-pad with zeros if the last segment is less than 6 bits
        const paddedBits = bits.padEnd(6, '0');
        const index = parseInt(paddedBits, 2);
        base64 += BASE64_CHARS[index];
    }

    // Calculate padding. Base64 output length must be divisible by 4.
    while (base64.length % 4 !== 0) {
        base64 += '=';
    }

    return base64;
}

export function Base64ToUint8Array(base64String: string): Uint8Array {
    let str = base64String.replace(/=+$/, ''); // Remove padding characters
    let bytes = [];

    for (let i = 0, len = str.length; i < len; i += 4) {
        let bitString = '';
        for (let j = 0; j < 4 && i + j < len; ++j) {
            const char = str.charAt(i + j);
            const index = BASE64_CHARS.indexOf(char);
            if (index !== -1) {
                bitString += index.toString(2).padStart(6, '0');
            }
        }

        for (let k = 0; k < bitString.length; k += 8) {
            if (k + 8 <= bitString.length) {
                const byte = bitString.substring(k, k + 8);
                bytes.push(parseInt(byte, 2));
            }
        }
    }

    return new Uint8Array(bytes);
}

export function Base64ToHexString(base64String: string): string {
    return Uint8ArrayToHex(Base64ToUint8Array(base64String));
}


export function HexStringToBase64(hexString: string): string {
    return Uint8ArrayToBase64(HexStringToUint8Array(hexString));
}


export function HexStringToUint8Array(hexString: string): Uint8Array {
    // Remove the '0x' prefix if present
    if (hexString.startsWith('0x')) {
        hexString = hexString.slice(2);
    }

    // Ensure the hex string has an even length
    if (hexString.length % 2 !== 0) {
        throw new Error('Invalid hex string');
    }

    // Create a Uint8Array with the appropriate length
    const byteArray = new Uint8Array(hexString.length / 2);

    // Convert each pair of hex characters to a byte
    for (let i = 0; i < hexString.length; i += 2) {
        byteArray[i / 2] = parseInt(hexString.substr(i, 2), 16);
    }

    return byteArray;
}

export function Uint8ArrayToHex(uint8Array: Uint8Array): string {
    return "0x" + uint8Array.reduce((str, byte) => str + byte.toString(16).padStart(2, '0'), '');
}