// @expect: {"b":"s","a":1} { b: 's', a: 1 } b,a
// @expect: {"b":"s","a":1,"c":true} b,a,c
// @expect: {"c":false,"a":2,"b":"t"} c,a,b
// @expect: {"y":1,"z":2}
// @expect: {"b":1,"a":2}
// @expect: {"n":{"q":"x","p":[1,2]},"m":null}
// JSON.stringify writes an object's properties in insertion order, like
// Object.keys and console.log, whatever order its declared type lists.
type T = { a: number; b: string; c?: boolean };
const x: T = { b: "s", a: 1 };
console.log(JSON.stringify(x), x, Object.keys(x).join(","));
x.c = true;
console.log(JSON.stringify(x), Object.keys(x).join(","));
const y: T = { c: false, a: 2, b: "t" };
console.log(JSON.stringify(y), Object.keys(y).join(","));
interface U {
  z: number;
  y: number;
}
const u: U = { y: 1, z: 2 };
console.log(JSON.stringify(u));
class K {
  b = 1;
  a = 2;
}
console.log(JSON.stringify(new K()));
type Inner = { p: number[]; q: string };
type Outer = { m: number | null; n: Inner };
const nested: Outer = { n: { q: "x", p: [1, 2] }, m: null };
console.log(JSON.stringify(nested));
