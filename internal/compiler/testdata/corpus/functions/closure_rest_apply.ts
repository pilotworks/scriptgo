// Closures with a rest parameter collect their trailing arguments, whether
// called directly, through a function-typed parameter, or with apply.
const join = (...parts: unknown[]) => parts.join("+");
const tag = (label: string, ...rest: string[]) => `${label}:${rest.join(",")}`;
console.log(join(1, 2, 3), tag("a"), tag("b", "x", "y"), tag("c", ...["p", "q"]));

function sum(...xs: number[]): number { return xs.reduce((a, b) => a + b, 0); }
const alias = sum;
function hof(fn: (...xs: number[]) => number) { return fn(4, 5, 6); }
console.log(alias(1, 2, 3), alias(), hof((...xs) => xs[2]), hof(sum));

// apply with a runtime-length array passes every element.
const xs: unknown[] = [1, 2, 3, 4, 5, 6];
function six(a: unknown, b: unknown, c: unknown, d: unknown, e: unknown, f: unknown): string { return [a, b, c, d, e, f].join(","); }
const few: unknown[] = ["ada"];
console.log(Reflect.apply(six, null, xs), Reflect.apply(six, null, few));
console.log(join.apply(null, xs), join.call(null, "u", "v"), Reflect.apply(join, null, few));
console.log([1, undefined, null, "z"].join("-"));
// @expect: 1+2+3 a: b:x,y c:p,q
// @expect: 6 0 6 15
// @expect: 1,2,3,4,5,6 ada,,,,,
// @expect: 1+2+3+4+5+6 u+v ada
// @expect: 1---z
