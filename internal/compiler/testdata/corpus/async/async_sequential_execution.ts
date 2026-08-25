// @expect: Step 1: 10
// @expect: Step 2: 20
// @expect: Step 3: 30
// @expect: Step 4: 40
// @expect: Total: 100
async function step(n: number): Promise<number> {
    return n * 10;
}

async function runSteps() {
    let total = 0;
    for (let i = 1; i <= 4; i++) {
        const val = await step(i);
        total += val;
        console.log("Step " + i + ": " + val);
    }
    console.log("Total: " + total);
}

runSteps();
