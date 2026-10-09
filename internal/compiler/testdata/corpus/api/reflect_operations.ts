// Reflect operations act on the target and report success as booleans.
// @expect: 2 5 7 true
// @expect: 6 15
// @expect: true 2 true
// @expect: true false false true 9
// @expect: false 1 false
// @expect: false 1 true 3
// @expect: true false
// @expect: true
// @expect: 4 true
class Point { constructor(public x: number, public y: number) {} sum() { return this.x + this.y; } }
const p = Reflect.construct(Point, [2, 5]);
console.log(p.x, p.y, p.sum(), p instanceof Point);
function add(a: number, b: number, c: number) { return a + b + c; }
const args: [number, number, number] = [1, 2, 3];
console.log(Reflect.apply(add, undefined, args), Reflect.apply(add, undefined, [4, 5, 6]));
const o: Record<string, number> = { a: 1 };
console.log(Reflect.set(o, "b", 2), o["b"], Reflect.isExtensible(o));
console.log(Reflect.preventExtensions(o), Reflect.isExtensible(o), Reflect.set(o, "c", 3), Reflect.set(o, "a", 9), o["a"]);
const f = Object.freeze({ q: 1 }) as Record<string, number>;
console.log(Reflect.set(f, "q", 2), f["q"], Reflect.deleteProperty(f, "q"));
const g = Object.freeze({ q: 1 });
const h = { q: 1 };
console.log(Reflect.defineProperty(g, "q", { value: 3 }), g.q, Reflect.defineProperty(h, "q", { value: 3 }), h.q);
const d: Record<string, number> = { z: 1 };
console.log(Reflect.deleteProperty(d, "z"), "z" in d);
console.log(Reflect.getPrototypeOf(p) === Object.getPrototypeOf(new Point(0, 0)));
const desc = Reflect.getOwnPropertyDescriptor({ k: 4 }, "k");
console.log(desc?.value, desc?.writable);
