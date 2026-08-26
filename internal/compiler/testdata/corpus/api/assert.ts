// ScriptGo Corpus: Assert Standard Builtin APIs
// Consolidated test suite with inline assertions covering all 27 official Node.js assert APIs.

import assert, {
    ok,
    equal,
    notEqual,
    strictEqual,
    notStrictEqual,
    deepEqual,
    notDeepEqual,
    deepStrictEqual,
    notDeepStrictEqual,
    partialDeepStrictEqual,
    throws,
    doesNotThrow,
    ifError,
    fail,
    match,
    doesNotMatch,
    rejects,
    doesNotReject,
    AssertionError,
    CallTracker,
    Assert
} from "node:assert";

// @api: assert(value[, message])
// @expect: assert passed
assert(true, "should be truthy");
console.log("assert passed");

// @api: assert.ok(value[, message])
// @expect: ok passed
ok(1, "1 is truthy");
console.log("ok passed");

// @api: assert.equal(actual, expected[, message])
// @expect: equal passed
equal(1, "1");
console.log("equal passed");

// @api: assert.notEqual(actual, expected[, message])
// @expect: notEqual passed
notEqual(1, 2);
console.log("notEqual passed");

// @api: assert.strictEqual(actual, expected[, message])
// @expect: strictEqual passed
strictEqual(42, 42);
console.log("strictEqual passed");

// @api: assert.notStrictEqual(actual, expected[, message])
// @expect: notStrictEqual passed
notStrictEqual(1, "1");
console.log("notStrictEqual passed");

// @api: assert.deepEqual(actual, expected[, message])
// @expect: deepEqual passed
deepEqual([1, 2], [1, 2]);
console.log("deepEqual passed");

// @api: assert.notDeepEqual(actual, expected[, message])
// @expect: notDeepEqual passed
notDeepEqual([1, 2], [1, 3]);
console.log("notDeepEqual passed");

// @api: assert.deepStrictEqual(actual, expected[, message])
// @expect: deepStrictEqual passed
deepStrictEqual([1, 2], [1, 2]);
console.log("deepStrictEqual passed");

// @api: assert.notDeepStrictEqual(actual, expected[, message])
// @expect: notDeepStrictEqual passed
notDeepStrictEqual({ a: 1 }, { b: 1 });
console.log("notDeepStrictEqual passed");

// @api: assert.partialDeepStrictEqual(actual, expected[, message])
// @expect: partialDeepStrictEqual passed
partialDeepStrictEqual({ a: 1, b: 2, c: 3 }, { a: 1, b: 2 });
console.log("partialDeepStrictEqual passed");

// @api: assert.throws(fn[, error][, message])
// @expect: throws passed
throws(() => {
    throw new Error("expected failure");
});
console.log("throws passed");

// @api: assert.doesNotThrow(fn[, error][, message])
// @expect: doesNotThrow passed
doesNotThrow(() => {
    const x = 1 + 1;
});
console.log("doesNotThrow passed");

// @api: assert.ifError(value)
// @expect: ifError passed
ifError(null);
ifError(undefined);
console.log("ifError passed");

// @api: assert.fail([message])
// @expect: fail caught
try {
    fail("explicit fail");
} catch (e) {
    console.log("fail caught");
}

// @api: assert.fail(actual, expected[, message[, operator[, stackStartFn]]])
// @expect: fail with args caught
try {
    fail(1, 2, "not equal", "!=");
} catch (e) {
    console.log("fail with args caught");
}

// @api: assert.match(string, regexp[, message])
// @expect: match passed
match("hello world", /world/);
console.log("match passed");

// @api: assert.doesNotMatch(string, regexp[, message])
// @expect: doesNotMatch passed
doesNotMatch("hello world", /foo/);
console.log("doesNotMatch passed");

// @api: assert.AssertionError
// @expect: AssertionError
// @expect: ERR_ASSERTION
const err = new AssertionError({ message: "custom assertion", actual: 1, expected: 2 });
console.log(err.name);
console.log(err.code);

// @api: assert.CallTracker
// @api: assert.CallTracker.calls
// @api: assert.CallTracker.getCalls
// @api: assert.CallTracker.report
// @api: assert.CallTracker.reset
// @api: assert.CallTracker.verify
// @expect: 2
// @expect: 0
// @expect: tracker verified
const tracker = new CallTracker();
function trackedFn(n: number): number {
    return n * 2;
}
const tracked = tracker.calls(trackedFn, 2);
tracked(10);
tracked(20);
console.log(tracker.getCalls(trackedFn).length);
console.log(tracker.report().length);
tracker.verify();
tracker.reset();
console.log("tracker verified");

// @api: assert.Assert
// @expect: Assert class passed
const customAssert = new Assert({ diff: "simple", strict: true });
customAssert.ok(true);
customAssert.strictEqual(100, 100);
console.log("Assert class passed");

// @api: assert.rejects
// @expect: rejects passed
await rejects(Promise.reject(new Error("async error")));
console.log("rejects passed");

// @api: assert.doesNotReject
// @expect: doesNotReject passed
await doesNotReject(Promise.resolve("ok"));
console.log("doesNotReject passed");

