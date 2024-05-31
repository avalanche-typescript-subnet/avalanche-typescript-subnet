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

    switch (payload[0]) {
        case 0: // ACTION_READ
            console.log(`ACTION_READ`)
            const dataRead: Record<string, number> = JSON.parse(
                decoder.decode(
                    getRawValue(0, 10)
                )
            );
            console.log(dataRead[actorString]
                ? `Balance of ${actorString.slice(0, 4)}...: ${dataRead[actorString]}`
                : `No balance for ${actorString.slice(0, 4)}..., but here are balances: ${Object.keys(dataRead).map(key => `${key.slice(0, 4)}...: ${dataRead[key]}`).join(', ')}`
            )

            return encoder.encode(JSON.stringify(dataRead[actorString] || 0));
        case 1: // ACTION_INCREMENT
            console.log(`ACTION_INCREMENT value=${value}`)
            const dataIncrement: Record<string, number> = JSON.parse(
                decoder.decode(
                    getRawValue(0, 10)
                )
            );

            if (!dataIncrement[actorString]) {
                dataIncrement[actorString] = 0;
            }
            dataIncrement[actorString] += value;
            setRawValue(0, 10, encoder.encode(JSON.stringify(dataIncrement)));

            console.log(`New balances: ${Object.keys(dataIncrement).map(key => `${key.slice(0, 4)}...: ${dataIncrement[key]}`).join(', ')}`)

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

