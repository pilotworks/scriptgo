// @parity-runner: tsc-node22
// ScriptGo Corpus: TypeScript 5.2 / ES2024 Explicit Resource Management (using & await using)
// Consolidated test suite verifying lexical disposal, LIFO order, and early return disposal.

class Resource {
    name: string;
    constructor(name: string) {
        this.name = name;
        console.log("created: " + name);
    }
    [Symbol.dispose]() {
        console.log("disposed: " + this.name);
    }
}

function testBlockScope() {
    console.log("entering block");
    {
        using r1 = new Resource("r1");
        using r2 = new Resource("r2");
        console.log("inside block");
    }
    console.log("exited block");
}

// @expect: entering block
// @expect: created: r1
// @expect: created: r2
// @expect: inside block
// @expect: disposed: r2
// @expect: disposed: r1
// @expect: exited block
testBlockScope();

function testEarlyReturn(shouldReturn: boolean): string {
    using rEarly = new Resource("rEarly");
    if (shouldReturn) {
        console.log("returning early");
        return "early_result";
    }
    console.log("normal exit");
    return "normal_result";
}

// @expect: created: rEarly
// @expect: returning early
// @expect: disposed: rEarly
// @expect: result: early_result
const res = testEarlyReturn(true);
console.log("result: " + res);
