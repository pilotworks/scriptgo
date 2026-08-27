// @expect: Attempt 1 failed: Transient error
// @expect: Attempt 2 failed: Transient error
// @expect: Attempt 3 succeeded: Recovered!
// @expect: Final status: Recovered!
async function unreliableOperation(attempt: number): Promise<string> {
    if (attempt < 3) {
        throw new Error("Transient error");
    }
    return "Recovered!";
}

async function retryWorker(): Promise<string> {
    for (let attempt = 1; attempt <= 3; attempt++) {
        try {
            const res = await unreliableOperation(attempt);
            console.log("Attempt " + attempt + " succeeded: " + res);
            return res;
        } catch (e) {
            const err = e as Error;
            console.log("Attempt " + attempt + " failed: " + err.message);
        }
    }
    return "Failed all";
}

async function main() {
    const status = await retryWorker();
    console.log("Final status: " + status);
}

main();
