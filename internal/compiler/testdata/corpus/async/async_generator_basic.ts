// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 40
async function* asyncRange(start: number, end: number) {
    for (let i = start; i <= end; i++) {
        yield i * 10;
    }
}

async function main() {
    for await (const val of asyncRange(1, 4)) {
        console.log(val);
    }
}

main();
