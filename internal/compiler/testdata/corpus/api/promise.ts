// @api: Promise.all
// @api: Promise.allSettled
// @api: Promise.any
// @api: Promise.race
// @api: Promise.reject
// @api: Promise.resolve
// @api: Promise.withResolvers
// @api: Promise.prototype.then
// @api: Promise.prototype.catch
// @api: Promise.prototype.finally
// @expect: resolved: 42
// @expect: 1
// @expect: 2
export {};

async function testPromise() {
    const p = Promise.resolve(42);
    const val = await p;
    console.log("resolved: " + val);

    const all = await Promise.all([Promise.resolve(1), Promise.resolve(2)]);
    console.log(all[0]);
    console.log(all[1]);
}

await testPromise();
