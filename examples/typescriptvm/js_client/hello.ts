import { TsChainClient } from "../runtime/js_sdk/src/client"

async function start() {
    const response = await fetch('http://localhost:12353/v1/control/status', {
        method: 'POST'
    });
    const data = await response.json();

    const uri = data.clusterInfo.nodeInfos.node1.uri;
    const chainId = Object.keys(data.clusterInfo.customChains)[0];
    const nodeUrl = `${uri}/ext/bc/${chainId}`;
    const client = new TsChainClient(nodeUrl);

    const balance = await client.balance("morpheus1qrzvk4zlwj9zsacqgtufx7zvapd3quufqpxk5rsdd4633m4wz2fdjk97rwu");
    console.log('balance is', balance);
}

start().then(() => process.exit(0)).catch(e => {
    console.error(e);
    process.exit(1);
});
