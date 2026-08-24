// ScriptGo Corpus: Language: Async, Promises & Generators
// Consolidated test suite with inline assertions.

// @expect: 150
// @expect: 10
// @expect: 100
// @expect: 200
// @expect: 20
// @expect: 1
// @expect: 2
// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 60
// @expect: 10
// @expect: 10
// @expect: false
// @expect: 20
// @expect: false
// @expect: 30
// @expect: false
// @expect: true
// @expect: 100
// @expect: 200
// @expect: 1
// @expect: 2
// @expect: 10
// @expect: 20
// @expect: 3
// @expect: 42

// --- Context Case: language_async_await ---
const p_async_await_0 = Promise.resolve(100);
const val_async_await_0 = await p_async_await_0;
console.log(val_async_await_0 + 50);
export {};

// --- Context Case: language_async_generators ---
function* subGen() {
    yield 100;
    yield 200;
}

function* mainGen() {
    yield 10;
    yield* subGen();
    yield 20;
}

for (const n of mainGen()) {
    console.log(n);
}

async function* asyncNumbers() {
    yield 1;
    yield 2;
}

for await (const x of asyncNumbers()) {
    console.log(x);
}

export {};

// --- Context Case: language_for_await_of ---
let sum_for_await_of_4 = 0;
const promises_for_await_of_4 = [Promise.resolve(10), Promise.resolve(20), Promise.resolve(30)];
for await (const x of promises_for_await_of_4) {
    sum_for_await_of_4 += x;
    console.log(x);
}
console.log(sum_for_await_of_4);

let numSum_for_await_of_4 = 0;
const nums_for_await_of_4 = [1, 2, 3, 4];
for await (const n of nums_for_await_of_4) {
    numSum_for_await_of_4 += n;
}
console.log(numSum_for_await_of_4);
export {};

// --- Context Case: language_generators ---
function* numbers() {
    yield 10;
    yield 20;
    yield 30;
}

const g_generators_5 = numbers();
const r1_generators_5 = g_generators_5.next();
console.log(r1_generators_5.value);
console.log(r1_generators_5.done);

const r2_generators_5 = g_generators_5.next();
console.log(r2_generators_5.value);
console.log(r2_generators_5.done);

const r3_generators_5 = g_generators_5.next();
console.log(r3_generators_5.value);
console.log(r3_generators_5.done);

const r4_generators_5 = g_generators_5.next();
console.log(r4_generators_5.done);

function* countTo(a: number, b: number) {
    yield a;
    yield b;
}

for (const n of countTo(100, 200)) {
    console.log(n);
}

// --- Context Case: language_async_microtask ---
console.log(1);
queueMicrotask(() => {
    console.log(3);
});
console.log(2);

// --- Context Case: language_async_promise ---
console.log(10);
const p_async_promise_3 = Promise.resolve(42);
p_async_promise_3.then((val) => {
    console.log(val);
});
console.log(20);

