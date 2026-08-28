// @expect: 20
// @expect: 10
// @expect: 42
// @expect: hello
// @expect: 10 20 30
// @expect: 1 20 300 400
// @expect: 1,10,2,20,3,30
// @expect: 3,6,9
// 1. Lexical block scope variable shadowing and isolation
const outerX: number = 10;
{
    const outerX: number = 20;
    console.log(outerX);
}
console.log(outerX);

const message: string = "hello";
{
    const message: number = 42;
    console.log(message);
}
console.log(message);

// 2. Object.fromEntries with multiple key-value tuple pairs
const pairs: [string, number][] = [["alpha", 10], ["beta", 20], ["gamma", 30]];
const converted: { alpha?: number; beta?: number; gamma?: number } = Object.fromEntries(pairs);
console.log(converted.alpha, converted.beta, converted.gamma);

// 3. Object.assign merging across multiple sequential sources
const baseObj = { a: 1, b: 2 };
const delta1 = { b: 20, c: 30 };
const delta2 = { c: 300, d: 400 };

const merged = Object.assign(baseObj, delta1, delta2);
console.log(merged.a, merged.b, merged.c, merged.d);

// 4. Array.prototype.flatMap and Array.from with mapping callback
const numbers: number[] = [1, 2, 3];
const expanded: number[] = numbers.flatMap((x: number) => [x, x * 10]);
console.log(expanded.join(","));

const tripled: number[] = Array.from([1, 2, 3], (x: number) => x * 3);
console.log(tripled.join(","));
