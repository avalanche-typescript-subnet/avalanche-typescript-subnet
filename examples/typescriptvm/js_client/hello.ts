import { TsChainClient, extractNodeUrlFromNetworkRunner } from "../runtime/js_sdk/src/client"

async function start() {
    const nodeUrl = await extractNodeUrlFromNetworkRunner();
    const client = new TsChainClient(nodeUrl);

    const balance = await client.balance("morpheus1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjk97rwu");
    console.log('balance is', balance);
}

start().then(() => process.exit(0)).catch(e => {
    console.error(e);
    process.exit(1);
});
