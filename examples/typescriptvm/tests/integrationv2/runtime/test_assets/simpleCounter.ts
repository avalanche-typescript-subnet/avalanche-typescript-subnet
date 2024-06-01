import { registerRawFunction, types, encoders } from "../../../../runtime/js_sdk"

const decoder = new TextDecoder();
const encoder = new TextEncoder();

registerRawFunction((payload: Uint8Array, actor: Uint8Array, getRawValue: types.GetRawValue, setRawValue: types.SetRawValue) => {
    //unpacking args
    const actorString = encoders.Uint8ArrayToBase64(actor);
    let value: number = 0;

    if (payload.length > 1) {
        value = payload[1];
    }

    let stringAtSlotZero: string
    let dataAtSlotZero: Record<string, number>

    switch (payload[0]) {
        case 0: // ACTION_READ
            console.log(`ACTION_READ`)
            stringAtSlotZero = decoder.decode(
                getRawValue(0, 10)
            )
            dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

            console.log(dataAtSlotZero[actorString]
                ? `Balance of ${actorString.slice(0, 4)}...: ${dataAtSlotZero[actorString]}`
                : `No balance for ${actorString.slice(0, 4)}..., but here are balances: ${Object.keys(dataAtSlotZero).map(key => `${key.slice(0, 4)}...: ${dataAtSlotZero[key]}`).join(', ')}`
            )

            return encoder.encode(JSON.stringify(dataAtSlotZero[actorString] || 0));
        case 1: // ACTION_INCREMENT
            console.log(`ACTION_INCREMENT value=${value}`)

            stringAtSlotZero = decoder.decode(getRawValue(0, 10))
            dataAtSlotZero = stringAtSlotZero === "" ? {} : JSON.parse(stringAtSlotZero);

            dataAtSlotZero[actorString] = (dataAtSlotZero[actorString] || 0) + value;

            setRawValue(0, 10, encoder.encode(JSON.stringify(dataAtSlotZero)));

            console.log(`New balances: ${Object.keys(dataAtSlotZero).map(key => `${key.slice(0, 4)}...: ${dataAtSlotZero[key]}`).join(', ')}`)

            return new Uint8Array();
        case 2: // ACTION_LOAD_CPU
            console.log(`ACTION_LOAD_CPU value=${value}`)
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
        case 3: //ACTION_WRITE_MANY_SLOTS
            console.log(`ACTION_WRITE_MANY_SLOTS value=${value}`)
            for (let i = 1; i <= value; i++) {
                setRawValue(
                    i,
                    6,
                    Uint8Array.from([i, i, i])
                );
            }
            return new Uint8Array();
        case 4: //ACTION_READ_MANY_SLOTS
            console.log(`ACTION_READ_MANY_SLOTS value=${value}`)
            let result = new Uint8Array();
            for (let i = 0; i < value; i++) {
                const fromSlot = getRawValue(i + 1, 6);
                let tempResult = new Uint8Array(result.length + fromSlot.length);
                tempResult.set(result);
                tempResult.set(fromSlot, result.length);
                result = tempResult;
            }
            return result;
        default:
            throw new Error("Invalid action code");
    }
});

