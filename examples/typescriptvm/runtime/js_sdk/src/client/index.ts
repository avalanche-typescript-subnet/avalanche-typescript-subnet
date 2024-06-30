

export class TsChainClient {
    constructor(private readonly url: string) {
    }

    public async balance(address: string): Promise<bigint> {
        const response = await fetch(this.url + "/morpheusapi", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                jsonrpc: "2.0",
                method: "morpheusvm.balance",
                params: { address },
                id: parseInt(String(Math.random()).slice(2))
            })
        });

        const json = await response.json();
        return BigInt(json.result.amount);
    }
}

export async function extractNodeUrlFromNetworkRunner(networkRunnerUrl: string = "http://localhost:12353"): Promise<string> {
    const response = await fetch(`${networkRunnerUrl}/v1/control/status`, {
        method: 'POST'
    });
    const data = await response.json();

    const uri = data.clusterInfo.nodeInfos.node1.uri;
    const chainId = Object.keys(data.clusterInfo.customChains)[0];
    const nodeUrl = `${uri}/ext/bc/${chainId}`;
    return nodeUrl;
}