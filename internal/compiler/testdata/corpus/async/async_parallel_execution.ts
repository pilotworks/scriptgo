// @expect: 2
// @expect: 4
// @expect: 6
async function fetchVal(n: number): Promise<number> {
    return n * 2;
}

async function main() {
    const promises = [fetchVal(1), fetchVal(2), fetchVal(3)];
    const results = await Promise.all(promises);
    for (const r of results) {
        console.log(r);
    }
}

main();
