// ScriptGo Corpus: node:test runner and mock APIs

import test, { describe, it, before, after, beforeEach, afterEach, mock } from "node:test";
import assert from "node:assert";

// @api: test
// @api: test.it
// @api: test.describe
// @api: test.suite
// @api: test.before
// @api: test.after
// @api: test.beforeEach
// @api: test.afterEach
// @api: test.mock

// @expect: ▶ Math Suite
// @expect:   ✔ addition test (0.000ms)
// @expect:   ✔ subtraction test (0.000ms)
// @expect:   ﹣ skipped test (0.000ms) # SKIP
// @expect:   ✔ todo test (0.000ms) # TODO
// @expect: ✔ Math Suite (0.000ms)
// @expect:   ✔ subtest 1 (0.000ms)
// @expect:   ✔ subtest 2 (0.000ms)
// @expect: ✔ Top level test with subtests (0.000ms)
// @expect: ✔ Mock function verification (0.000ms)
// @expect: ℹ tests 2
// @expect: ℹ suites 1
// @expect: ℹ pass 6
// @expect: ℹ fail 0
// @expect: ℹ cancelled 0
// @expect: ℹ skipped 1
// @expect: ℹ todo 1

process.env["NODE_TEST_DETERMINISTIC"] = "1";
let hookCounter = 0;

describe("Math Suite", () => {
    before(() => {
        hookCounter = 10;
    });

    beforeEach(() => {
        hookCounter += 1;
    });

    afterEach(() => {
        hookCounter += 1;
    });

    after(() => {
        assert.strictEqual(hookCounter, 16);
    });

    it("addition test", () => {
        assert.strictEqual(1 + 1, 2);
    });

    it("subtraction test", () => {
        assert.strictEqual(5 - 3, 2);
    });

    it.skip("skipped test", () => {
        throw new Error("Should not execute");
    });

    it.todo("todo test", () => {
        // Pending implementation
    });
});

test("Top level test with subtests", async (t) => {
    let subtestRun = 0;
    await t.test("subtest 1", () => {
        subtestRun += 1;
        assert.strictEqual(subtestRun, 1);
    });
    await t.test("subtest 2", () => {
        subtestRun += 1;
        assert.strictEqual(subtestRun, 2);
    });
    assert.strictEqual(subtestRun, 2);
});

test("Mock function verification", () => {
    const add = (a: number, b: number) => a + b;
    const mockedAdd = mock.fn(add);

    const r1 = mockedAdd(2, 3);
    assert.strictEqual(r1, 5);
    assert.strictEqual(mockedAdd.mock.callCount(), 1);

    mockedAdd.mock.mockImplementationOnce((a: number, b: number) => a * b);
    const r2 = mockedAdd(4, 5);
    assert.strictEqual(r2, 20);

    const r3 = mockedAdd(2, 3);
    assert.strictEqual(r3, 5);
    assert.strictEqual(mockedAdd.mock.callCount(), 3);
});
