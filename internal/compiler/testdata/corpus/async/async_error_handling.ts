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
    } catch (e: any) {
        console.log("Caught: " + e.message);
    }

    try {
        const bad = await riskyOperation(true);
        console.log(bad);
    } catch (e: any) {
        console.log("Caught error: " + e.message);
    } finally {
        console.log("Cleanup done");
    }
}

main();
