// @expect: caught inner
// @expect: finally 1
// @expect: finally 2
// @expect: Result 1: 20
// @expect: Result 2: A-B-C-D-E-F
function testFinallyReturnOverride(): number {
    try {
        try {
            throw new Error("inner error");
        } catch (e) {
            console.log("caught inner");
            return 10;
        } finally {
            console.log("finally 1");
            return 20; // overrides return 10
        }
    } finally {
        console.log("finally 2");
    }
}

function testNestedExceptionFlow(): string {
    let trace = "";
    try {
        trace += "A-";
        try {
            trace += "B-";
            throw new Error("fail");
        } catch (e) {
            trace += "C-";
            throw e; // rethrow
        } finally {
            trace += "D-";
        }
    } catch (e) {
        trace += "E-";
    } finally {
        trace += "F";
    }
    return trace;
}

console.log("Result 1: " + testFinallyReturnOverride());
console.log("Result 2: " + testNestedExceptionFlow());
