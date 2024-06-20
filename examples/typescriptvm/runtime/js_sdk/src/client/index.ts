

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