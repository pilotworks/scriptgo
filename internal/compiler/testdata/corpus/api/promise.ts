// ScriptGo Corpus: Promise Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: Promise.all
// @api: Promise.resolve
// @api: Promise.withResolvers
// @api: Promise.try
// @api: new Promise
// @api: Promise.constructor
// @expect: resolved: 42
// @expect: bool: true
// @expect: bigint: 123
// @expect: 1
// @expect: 2
// @expect: 100
// @expect: object
// @expect: constructed: 7
// @expect: try: 999
// @expect: withResolvers: 8
export {};

function getVal(): number {
    return 999;
}

async function testPromise() {
    const p = Promise.resolve(42);
    const val = await p;
    console.log("resolved: " + val);

    const boolValue = await Promise.resolve(true);
    console.log("bool: " + boolValue);
    const bigintValue = await Promise.resolve(123n);
    console.log("bigint: " + bigintValue.toString());

    const all = await Promise.all([Promise.resolve(1), Promise.resolve(2)]);
    console.log(all[0]);
    console.log(all[1]);

    const { promise, resolve } = Promise.withResolvers<number>();
    resolve(8);
    const withResolversValue = await promise;
    const resTry = await Promise.try(() => getVal());

    const pCreated = new Promise<number>((res) => {
        const dummy = true;
    });
    console.log(100);
    console.log(typeof pCreated);
    const constructed = await new Promise<number>((resolve) => {
        resolve(7);
    });
    console.log("constructed: " + constructed);
    console.log("try: " + resTry);
    console.log("withResolvers: " + withResolversValue);
}

await testPromise();
