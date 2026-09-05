// @expect: 1
async function findFirst(): Promise<number> {
    let i = 0;
    while (i < 3) {
        const stop = await Promise.resolve(i === 1);
        if (stop) {
            break;
        }
        i++;
    }
    return i;
}

async function main(): Promise<void> {
    console.log(await findFirst());
}

main();
