// @expect: === Advanced Control Flow & Scoping Test ===
// @expect: Labeled Loops: (0,0) (0,1) (0,2) (0,3) (1,0) (1,1) (2,0) (2,1)
// @expect: Switch Evaluation: 10:A, 9:A, 8:B, 7:B, 5:C, 2:F
// @expect: Closures in Loop: 0, 11, 22, 33, 44
// @expect: Normal: ret=SUCCESS, trace=try-start->try-end->finally-block
// @expect: Handled: ret=HANDLED, trace=try-start->throwing->caught:Failure->finally-block
// @expect: Rethrown: trace=try-start->throwing->caught:Failure->rethrowing->finally-block->outer-catch
// @expect: Omitted Catch Binding Handled: true
// @expect: Logical Assignments: a=42, b=10, e=200, f=0, c=default-c, d=default-d, g=existing-g
// @expect: Comma Operator: x=5, y=15
// @expect: For loop with commas: arr=[0, 8, 12, 12, 8], count=30

// Advanced Control Flow, Loops, Labeled Breaks/Continues, Closures, Scoping & Short-Circuiting

function testLabeledLoops(): void {
    let result = "";
    outerLoop: for (let i = 0; i < 4; i++) {
        innerLoop: for (let j = 0; j < 4; j++) {
            if (i === 1 && j === 2) {
                continue outerLoop;
            }
            if (i === 2 && j === 2) {
                break outerLoop;
            }
            result += `(${i},${j}) `;
        }
    }
    console.log("Labeled Loops:", result.trim());
}

function testNestedSwitchFallthrough(): void {
    function evaluateGrade(code: number): string {
        let category = "";
        switch (code) {
            case 10:
            case 9:
                category = "A";
                break;
            case 8:
            case 7:
                category = "B";
                break;
            case 6:
            case 5:
                category = "C";
                break;
            case 4:
            case 3:
                category = "D";
                break;
            default:
                category = "F";
                break;
        }
        return category;
    }

    const testScores = [10, 9, 8, 7, 5, 2];
    const categories = testScores.map(s => `${s}:${evaluateGrade(s)}`);
    console.log("Switch Evaluation:", categories.join(", "));
}

function testClosuresInLoops(): void {
    const funcs: (() => number)[] = [];
    for (let i = 0; i < 5; i++) {
        const factor = i * 10;
        funcs.push(() => factor + i);
    }

    const outputs: number[] = [];
    for (let j = 0; j < funcs.length; j++) {
        outputs.push(funcs[j]());
    }
    console.log("Closures in Loop:", outputs.join(", "));
}

function testTryCatchFinallyExecutionOrder(): void {
    const trace: string[] = [];

    function riskyOperation(shouldThrow: boolean, handled: boolean): string {
        try {
            trace.push("try-start");
            if (shouldThrow) {
                trace.push("throwing");
                throw new Error("Failure");
            }
            trace.push("try-end");
            return "SUCCESS";
        } catch (err) {
            const e = err as Error;
            trace.push(`caught:${e.message}`);
            if (!handled) {
                trace.push("rethrowing");
                throw err;
            }
            return "HANDLED";
        } finally {
            trace.push("finally-block");
        }
    }

    // Case 1: Normal execution
    const r1 = riskyOperation(false, true);
    console.log(`Normal: ret=${r1}, trace=${trace.join("->")}`);

    // Case 2: Handled error
    trace.length = 0;
    const r2 = riskyOperation(true, true);
    console.log(`Handled: ret=${r2}, trace=${trace.join("->")}`);

    // Case 3: Rethrown error with outer handler
    trace.length = 0;
    try {
        riskyOperation(true, false);
    } catch {
        trace.push("outer-catch");
    }
    console.log(`Rethrown: trace=${trace.join("->")}`);
}

function testOmittedCatchBinding(): void {
    let handled = false;
    try {
        throw new Error("Silent error");
    } catch {
        handled = true;
    }
    console.log("Omitted Catch Binding Handled:", handled);
}

function testLogicalAssignmentsAndShortCircuit(): void {
    let a = 0;
    let b = 10;
    let c: string | null = null;
    let d: string | undefined = undefined;

    // Logical OR assignment (||=)
    a ||= 42;
    b ||= 99;

    // Logical AND assignment (&&=)
    let e = 100;
    let f = 0;
    e &&= 200;
    f &&= 300;

    // Nullish Coalescing assignment (??=)
    c ??= "default-c";
    d ??= "default-d";
    let g: string | null = "existing-g";
    g ??= "ignored-g";

    console.log(`Logical Assignments: a=${a}, b=${b}, e=${e}, f=${f}, c=${c}, d=${d}, g=${g}`);
}

function testCommaOperatorAndComplexExpressions(): void {
    let x = 0;
    let y = (x += 5, x * 2, x + 10);
    console.log(`Comma Operator: x=${x}, y=${y}`);

    let count = 0;
    const arr: number[] = [];
    for (let i = 0, j = 10; i < 5; i++, j -= 2) {
        count += (arr.push(i * j), j);
    }
    console.log(`For loop with commas: arr=[${arr.join(", ")}], count=${count}`);
}

function main(): void {
    console.log("=== Advanced Control Flow & Scoping Test ===");
    testLabeledLoops();
    testNestedSwitchFallthrough();
    testClosuresInLoops();
    testTryCatchFinallyExecutionOrder();
    testOmittedCatchBinding();
    testLogicalAssignmentsAndShortCircuit();
    testCommaOperatorAndComplexExpressions();
}

main();
