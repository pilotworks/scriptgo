// @expect: K { b: 1, a: 2 }
// @expect: Empty {}
// @expect: Child { x: 1, y: [ 1, 2 ], inner: K { b: 1, a: 2 } }
// @expect: { k: K { b: 1, a: 2 }, list: [ K { b: 1, a: 2 } ] }
// @expect: [ K { b: 1, a: 2 } ]
// @expect: {"x":1,"y":[1,2],"inner":{"b":1,"a":2}}
// @expect: Map(1) { 'a' => K { b: 1, a: 2 } }
// @expect: [ [ 1, 2 ], [ 3 ] ] { n: [ [ 1, 2 ], [ 3 ] ] } {"n":[[1,2],[3]]}
// @expect: { b: [ true, false ] } {"b":[true,false]}
// @expect: [object Object]
// @expect: [object Object]
// @expect: P(p) P(p) xP(p)
// @expect: 1,2 [object Object] Error: e
// console.log prefixes class instances with their class name and prints
// nested, boolean and object arrays like Node's util.inspect. Template
// literals, String() and `+` concatenation apply ToString: plain objects
// become "[object Object]", arrays join their elements and a class's own
// toString() is called.
class K { b = 1; a = 2; }
class Empty {}
class Base { x = 1; }
class Child extends Base { y = [1, 2]; inner = new K(); }
const o = { k: new K(), list: [new K()] };
console.log(new K());
console.log(new Empty());
console.log(new Child());
console.log(o);
console.log([new K()]);
console.log(JSON.stringify(new Child()));
const m = new Map<string, K>([["a", new K()]]);
console.log(m);
const n: number[][] = [[1, 2], [3]];
console.log(n, { n }, JSON.stringify({ n }));
const b = [true, false];
console.log({ b }, JSON.stringify({ b }));
class P { name = "p"; toString(): string { return "P(" + this.name + ")"; } }
const k = new K();
console.log(`${k}`);
console.log(String(k));
console.log(`${new P()}`, String(new P()), "x" + new P());
console.log(`${[1, 2]}`, String({ a: 1 }), `${new Error("e")}`);
