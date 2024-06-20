import { TsChainClient } from "../runtime/js_sdk/src/client"
import * as fs from "node:fs"

const nodeUrl = fs.readFileSync(__dirname + "/../.morpheus-cli/nodeUrl.txt", "utf-8")
const client = new TsChainClient(nodeUrl)

async function start() {
    const balance = await client.balance("morpheus1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjk97rwu")
    console.log(balance)
}

start().then(() => process.exit(0)).catch(e => {
    console.error(e)
    process.exit(1)
})