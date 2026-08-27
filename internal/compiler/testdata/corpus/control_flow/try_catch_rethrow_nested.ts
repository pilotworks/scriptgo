// @expect: inner catch: level 1 failed
// @expect: inner finally executed
// @expect: outer catch: wrapped: level 1 failed
// @expect: outer finally executed
// @expect: completed
function riskyOperation(): void {
    try {
        try {
            throw new Error("level 1 failed");
        } catch (e: unknown) {
            const err = e as Error;
            console.log("inner catch: " + err.message);
            throw new Error("wrapped: " + err.message);
        } finally {
            console.log("inner finally executed");
        }
    } catch (e: unknown) {
        const err = e as Error;
        console.log("outer catch: " + err.message);
    } finally {
        console.log("outer finally executed");
    }
}

riskyOperation();
console.log("completed");
