// @expect: prefix
// @expect: after-false
// @expect: true
// @expect: prefix
// @expect: after-true
// @expect: true

async function branch(value: boolean): Promise<void> {
    console.log("prefix");
    if (value) {
        await Promise.resolve(undefined);
        console.log("after-true");
    } else {
        console.log("after-false");
    }
}

async function main(): Promise<void> {
    console.log(await branch(false) === undefined);
    console.log(await branch(true) === undefined);
}

main();
