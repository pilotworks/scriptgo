// @expect: 42
const multiplyAsync = async (a: number, b: number): Promise<number> => {
    return a * b;
};

async function main() {
    const res = await multiplyAsync(6, 7);
    console.log(res);
}

main();
