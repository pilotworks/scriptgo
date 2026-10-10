// Closures take any number of parameters: the first four travel in the
// closure ABI's registers, the rest through the runtime's extra-argument slot.
const five = (a: number, b: number, c: number, d: number, e: number) => a + b + c + d + e;
const six = (a: string, b: string, c: string, d: string, e: string, f: string) => [a, b, c, d, e, f].join("-");
console.log(five(1, 2, 3, 4, 5), six("a", "b", "c", "d", "e", "f"));

function apply7(fn: (a: number, b: number, c: number, d: number, e: number, f: number, g: number) => number): number {
  return fn(1, 2, 3, 4, 5, 6, 7);
}
let base = 100;
console.log(apply7((a, b, c, d, e, f, g) => base + a * b * c + d + e + f * g));

// A closure called inside another's extra arguments keeps both lists apart.
const outer = (a: number, b: number, c: number, d: number, e: number) => five(e, d, c, b, a) * e;
console.log(outer(1, 1, 1, 1, 2));

// Array callbacks receive the array as their third argument.
console.log([1, 2, 3].map((v, i, arr) => v * arr.length + i), [4, 5].reduce((s, v, i, arr) => s + v * arr.length, 0));
// @expect: 15 a-b-c-d-e-f
// @expect: 157
// @expect: 12
// @expect: [ 3, 7, 11 ] 18
