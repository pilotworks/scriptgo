// @expect: Map(4) { 1 => 'one', 2.5 => 'two-half', NaN => 'nan', 0 => 'zero' } 4 one zero nan true
// @expect: 2 one
// @expect: 3.5 two-half
// @expect: NaN nan
// @expect: 1 zero
// @expect: [ 2, 5, NaN, 0 ]
// @expect: number 1 one
// @expect: number 2.5 two-half
// @expect: number NaN nan
// @expect: number 0 zero
// @expect: true false 3
// @expect: Map(4) { 1 => 'one', NaN => 'nan', 0 => 'zero', 4 => 'four' } 3
// @expect: Map(0) {} 0
// @expect: 2 10 20 Map(2) { 1 => 10, '1' => 20 }
// @expect: [ 1, '1' ] true true false Map(2) { 1 => 'n', '1' => 's' }
// @expect: number 1
// @expect: string 1
// @expect: Map(2) { 1n => 'a', 2n => 'b' } a
// @expect: 2 obj Map(2) { { id: 1 } => 'obj', { id: 1 } => 'other' }
// @expect: yes false Map(1) { true => 'yes' }
// @expect: 7 false Map(1) { Symbol(tag) => 7 }
// @expect: true Map(1) { 10 => true }
// @expect: x 2
// @expect: y 1
// @expect: Map(2) { 'x' => 2, 'y' => 1 } [ 'x', 'y' ] [ [ 'x', 2 ], [ 'y', 1 ] ]
// @expect: undefined true undefined
// @expect: undefined true dflt
// @expect: undefined true 0
// Map keys keep their JavaScript type and compare with SameValueZero: the
// number 1 and the string "1" are distinct keys, NaN finds NaN, -0 is +0,
// object and symbol keys compare by identity, and console.log prints each
// key in its own form. get() of an absent key is undefined.
const m = new Map<number, string>();
m.set(1, "one"); m.set(2.5, "two-half"); m.set(NaN, "nan"); m.set(-0, "zero");
console.log(m, m.size, m.get(1), m.get(0), m.get(NaN), m.has(2.5));
for (const [k, v] of m) console.log(k + 1, v);
console.log([...m.keys()].map((k) => k * 2));
m.forEach((v, k) => console.log(typeof k, k, v));
console.log(m.delete(2.5), m.delete(2.5), m.size);
const copy = new Map(m);
copy.set(4, "four");
console.log(copy, m.size);
m.clear();
console.log(m, m.size);

const mixed = new Map<string | number, number>();
mixed.set(1, 10); mixed.set("1", 20);
console.log(mixed.size, mixed.get(1), mixed.get("1"), mixed);
const pairs = new Map<string | number, string>([[1, "n"], ["1", "s"]]);
console.log([...pairs.keys()], pairs.has(1), pairs.has("1"), pairs.has(2), pairs);
for (const k of pairs.keys()) console.log(typeof k, k);

const big = new Map<bigint, string>([[1n, "a"]]);
big.set(2n, "b");
console.log(big, big.get(1n));
const objKey = { id: 1 };
const byObject = new Map<object, string>();
byObject.set(objKey, "obj"); byObject.set({ id: 1 }, "other");
console.log(byObject.size, byObject.get(objKey), byObject);
const flags = new Map<boolean, string>([[true, "yes"]]);
console.log(flags.get(true), flags.has(false), flags);
const sym = Symbol("tag");
const bySymbol = new Map<symbol, number>([[sym, 7]]);
console.log(bySymbol.get(sym), bySymbol.has(Symbol("tag")), bySymbol);
const nested = { lookup: new Map<number, boolean>([[10, true]]) };
console.log(nested.lookup.get(10), nested.lookup);

const words = new Map<string, number>();
for (const w of ["x", "y", "x"]) words.set(w, (words.get(w) ?? 0) + 1);
words.forEach((count, word) => console.log(word, count));
console.log(words, [...words.keys()], [...words.entries()]);
const missing = words.get("zz");
console.log(missing, missing === undefined, typeof missing);
const strMap = new Map<string, string>();
const s = strMap.get("a");
console.log(s, s === undefined, strMap.get("a") ?? "dflt");
const arrMap = new Map<string, number[]>();
console.log(arrMap.get("q"), arrMap.get("q") === undefined, (arrMap.get("q") ?? []).length);
