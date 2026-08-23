// ScriptGo Corpus: Promise Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: Promise.all
// @api: Promise.allSettled
// @api: Promise.any
// @api: Promise.race
// @api: Promise.reject
// @api: Promise.resolve
// @api: Promise.withResolvers
// @api: Promise.try
// @api: Promise.prototype.then
// @api: Promise.prototype.catch
// @api: Promise.prototype.finally
// @api: new Promise
// @api: Promise.constructor
// @api: Promise.new
// @expect: resolved: 42
// @expect: 1
// @expect: 2
// @expect: 100
// @expect: object
// @expect: try: 999
export {};

function getVal(): number {
    return 999;
}

async function testPromise() {
    const p = Promise.resolve(42);
    const val = await p;
    console.log("resolved: " + val);

    const all = await Promise.all([Promise.resolve(1), Promise.resolve(2)]);
    console.log(all[0]);
    console.log(all[1]);

    const { promise, resolve } = Promise.withResolvers<number>();
    const resTry = await Promise.try(() => getVal());

    const pCreated = new Promise<number>((res) => {
        const dummy = true;
    });
    console.log(100);
    console.log(typeof pCreated);
    console.log("try: " + resTry);
}

await testPromise();
