// @expect: NaN undefined true false
// @expect: {"x":null}
// @expect: NaN undefined true false false
// @expect: a {"a":null}
// @expect: NaN NaN false { n: NaN }
// @expect: NaN [ { v: NaN } ]
// @expect: { a: NaN }
// @expect: undef num:NaN num:NaN num:1
// @expect: true true false
// @expect: false false NaN
// @expect: 7 true true undefined undefined
// @expect: NaN false { p: NaN }
// @expect: undefined false number object
// @expect: {"x":0} {"a":1}
// A NaN number is a value: storing it in a number field, parameter or
// variable never makes it read back as undefined or null, which number
// storage holds as reserved NaN payloads.
class P {
  x: number = 0;
  y?: number;
}
const p = new P();
p.x = NaN;
console.log(p.x, p.y, Number.isNaN(p.x), p.x === undefined);
console.log(JSON.stringify(p));
const o: { a: number; b?: number } = { a: NaN };
console.log(o.a, o.b, "a" in o, "b" in o, o.a === undefined);
console.log(Object.keys(o).join(","), JSON.stringify(o));
const q = { n: 0 / 0 };
console.log(q.n, String(q.n), q.n === undefined, q);
const arr = [{ v: NaN }];
console.log(arr[0].v, arr);
console.log(o);

function f(x?: number): string {
  return x === undefined ? "undef" : "num:" + x;
}
console.log(f(), f(NaN), f(0 / 0), f(1));
let v: number | null = null;
console.log(v === null, v == undefined, v === undefined);
v = NaN;
console.log(v === null, v == null, v ?? 5);
const e: { b?: number } = {};
console.log(e.b ?? 7, e.b === undefined, e.b == null, `${e.b}`, String(e.b));
const parsed = parseFloat("abc");
const box: { p?: number } = { p: parsed };
console.log(box.p, box.p === undefined, box);

const k: { padding?: number } = {};
const absent = typeof k.padding;
k.padding = NaN;
let maybe: number | null = null;
console.log(absent, absent === "number", typeof k.padding, typeof maybe);
class Q {
  x: number = 0;
  y?: number;
}
const plain: { a: number; b?: number } = { a: 1 };
console.log(JSON.stringify(new Q()), JSON.stringify(plain));
