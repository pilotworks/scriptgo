// @expect: Operation succeeded
// @expect: Caught error: Something went wrong
// @expect: Cleanup done
async function riskyOperation(fail: boolean): Promise<string> {
    if (fail) {
        throw new Error("Something went wrong");
    }
    return "Operation succeeded";
}

async function main() {
    try {
        const ok = await riskyOperation(false);
        console.log(ok);
    } catch (e: unknown) {
        const err = e as Error;
        console.log("Caught: " + err.message);
    }

    try {
        const bad = await riskyOperation(true);
        console.log(bad);
    } catch (e: unknown) {
        const err = e as Error;
        console.log("Caught error: " + err.message);
    } finally {
        console.log("Cleanup done");
    }
}

main();
