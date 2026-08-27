// @expect: Stage 1: 10
// @expect: Stage 2: 15
// @expect: Final: 15
async function stage1(input: number): Promise<number> {
    return input * 2;
}

async function stage2(input: number): Promise<number> {
    return input + 5;
}

async function runPipeline() {
    const raw = 5;
    const s1 = await stage1(raw);
    console.log("Stage 1: " + s1);
    const s2 = await stage2(s1);
    console.log("Stage 2: " + s2);
    console.log("Final: " + s2);
}

runPipeline();
