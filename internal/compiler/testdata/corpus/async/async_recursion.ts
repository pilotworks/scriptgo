// @expect: 120
async function asyncFactorial(n: number): Promise<number> {
    if (n <= 1) return 1;
    const prev = await asyncFactorial(n - 1);
    return n * prev;
}

async function main() {
    const res = await asyncFactorial(5);
    console.log(res);
}

main();
