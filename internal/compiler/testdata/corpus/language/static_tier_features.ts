// ScriptGo Corpus: Static Tier Features Test Suite
// Consolidated test suite with inline assertions.

// --- Context Case: math_and_number_apis ---
// @expect: 3
console.log(Math.cbrt(27));
// @expect: 27
console.log(Math.clz32(16));
// @expect: 12
console.log(Math.imul(3, 4));
// @expect: 0
console.log(Math.sinh(0));
// @expect: true
console.log(Number.isSafeInteger(9007199254740991));
// @expect: false
console.log(Number.isSafeInteger(9007199254740992));
// @expect: true
console.log(Number.EPSILON > 0);

// --- Context Case: string_and_unicode_apis ---
// @expect: A
console.log(String.fromCodePoint(65));
// @expect: 65
console.log("ABC".codePointAt(0));
// @expect: true
console.log("hello".isWellFormed());
// @expect: hello
console.log("hello".toWellFormed());

// --- Context Case: object_and_clone_apis ---
// @expect: 42
let clone_target = { a: 42, b: "hello" };
let cloned = structuredClone(clone_target);
console.log(cloned.a);

// --- Context Case: promise_and_async_apis ---
// @expect: true
let resolvers = Promise.withResolvers();
console.log(resolvers.promise !== null);

async function testAsyncStaticFeatures(): Promise<number> {
    let p1 = Promise.resolve(10);
    let p2 = Promise.resolve(20);
    let anyRes = await Promise.any([p1, p2]);
    console.log(anyRes);
    let settled = await Promise.allSettled([p1, p2]);
    console.log(settled[0].status);
    return anyRes;
}
// @expect: 10
// @expect: fulfilled
testAsyncStaticFeatures();

// --- Context Case: auto_accessors ---
class AccessorTest {
    accessor count: number = 5;
}
let acc = new AccessorTest();
// @expect: 5
console.log(acc.count);
acc.count = 15;
// @expect: 15
console.log(acc.count);

// --- Context Case: namespaces ---
namespace Calculator {
    export function multiply(a: number, b: number): number {
        return a * b;
    }
}
// @expect: 42
console.log(Calculator.multiply(6, 7));
