// Closures with a rest parameter collect the arguments the call passed,
// whether called directly, through a function-typed parameter, with
// call/apply, or by a built-in such as map, reduce, sort or replace.
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

// Built-ins pass their callbacks a fixed argument list; the rest parameter
// holds exactly those arguments.
console.log([10, 20].map((...a: unknown[]) => a.length), [1, 2, 3].reduce((...a: unknown[]) => (a[0] as number) + (a[1] as number) + a.length, 0));
[5].forEach((v: number, ...rest: unknown[]) => console.log(v, rest.length));
console.log([3, 1, 2].sort((...p: number[]) => p[0] - p[1]), "a-b".replace(/(\w)-(\w)/, (...m: string[]) => m.slice(0, 3).join("|") + "/" + m.length));
// Passed undefined arguments count; absent ones do not.
const count = (...xs: (number | undefined)[]) => xs.length;
console.log(count(), count(1, undefined), count(undefined, undefined, undefined, undefined, undefined));
// @expect: 1+2+3 a: b:x,y c:p,q
// @expect: 6 0 6 15
// @expect: 1,2,3,4,5,6 ada,,,,,
// @expect: 1+2+3+4+5+6 u+v ada
// @expect: 1---z
// @expect: [ 3, 3 ] 18
// @expect: 5 2
// @expect: [ 1, 2, 3 ] a-b|a|b/5
// @expect: 0 2 5
