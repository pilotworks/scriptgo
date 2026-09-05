// @expect: cleanup
// @expect: recovered: 7

async function run(): Promise<number> {
    try {
        await Promise.resolve(undefined);
        throw new Error("failure");
    } catch (error) {
        return 7;
    } finally {
        console.log("cleanup");
    }
}

run().then((value: number) => console.log("recovered: " + value));
