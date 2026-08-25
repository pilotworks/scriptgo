// @expect: ITEM:1
// @expect: ITEM:2
// @expect: ITEM:3
class AsyncService {
    private prefix: string;

    constructor(prefix: string) {
        this.prefix = prefix;
    }

    async fetchData(id: number): Promise<string> {
        return this.prefix + ":" + id;
    }

    async processAll(ids: number[]): Promise<string[]> {
        const results: string[] = [];
        for (const id of ids) {
            const data = await this.fetchData(id);
            results.push(data);
        }
        return results;
    }
}

async function main() {
    const service = new AsyncService("ITEM");
    const items = await service.processAll([1, 2, 3]);
    for (const item of items) {
        console.log(item);
    }
}

main();
