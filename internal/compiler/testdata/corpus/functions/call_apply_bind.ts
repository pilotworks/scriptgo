// Function.prototype.call/apply/bind invoke the function with the given
// arguments (thisArg is not observable without a `this` parameter).

function add(a: number, b: number, c: number) { return a + b + c; }
const args: [number, number, number] = [1, 2, 3];

// @expect: 15
console.log(add.apply(undefined, [4, 5, 6]));
// @expect: 6
console.log(add.apply(undefined, args));
// @expect: 3
console.log(add.call(undefined, 1, 1, 1));

const mul = (a: number, b: number) => a * b;
// @expect: 12
console.log(mul.apply(null, [3, 4]));
// @expect: 20
const bound = mul.bind(null);
console.log(bound(4, 5));

// A runtime-length unknown[] passes its elements (up to the closure ABI's
// four arguments); missing ones are undefined.
function show(a: unknown, b: unknown): string { return String(a) + "/" + String(b); }
const values: unknown[] = ["ada"];
// @expect: ada/undefined
console.log(Reflect.apply(show, undefined, values));
