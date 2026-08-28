// @expect: from_finally
// @expect: suppressed_by_finally
// @expect: try_0,finally_0,try_1,finally_1
// @expect: 10
// Complex try-catch-finally control flow semantics

// 1. Finally return overrides try return
function testFinallyOverride(): string {
    try {
        return "from_try";
    } finally {
        return "from_finally";
    }
}
console.log(testFinallyOverride());

// 2. Finally return overrides catch re-throw
function testFinallySuppressThrow(): string {
    try {
        throw new Error("catch_me");
    } catch (e) {
        throw new Error("rethrow_from_catch");
    } finally {
        return "suppressed_by_finally";
    }
}
console.log(testFinallySuppressThrow());

// 3. Nested try-finally with loop break
let trace: string[] = [];
for (let i = 0; i < 3; i++) {
    try {
        trace.push(`try_${i}`);
        if (i === 1) {
            break;
        }
    } finally {
        trace.push(`finally_${i}`);
    }
}
console.log(trace.join(","));

// 4. Primitive return value preservation vs finally mutation
function testValuePreservation(): number {
    let x = 10;
    try {
        return x;
    } finally {
        x = 99;
    }
}
console.log(testValuePreservation());
