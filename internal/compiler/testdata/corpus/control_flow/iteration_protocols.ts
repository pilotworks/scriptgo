// @expect: k a
// @expect: k b
// @expect: v 1
// @expect: v 2
// @expect: e a 1
// @expect: e b 2
// @expect: m a 1
// @expect: m b 2
// @expect: a,b [ [ 'a', 1 ], [ 'b', 2 ] ] [ [ 'a', 1 ], [ 'b', 2 ] ]
// @expect: x x
// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 0,1,2 5,9
// @expect: 7 false 8 true
// @expect: p
// @expect: q
// Iteration protocols: Map and Set iterators yield typed entries, a class
// iterates through its *[Symbol.iterator]() generator, generator methods,
// generator closures and values typed only Generator<T> all advance through
// next(), and a yield may read any parameter expression.
const m = new Map<string, number>([
  ["a", 1],
  ["b", 2],
]);
for (const k of m.keys()) console.log("k", k);
for (const v of m.values()) console.log("v", v);
for (const [k, v] of m.entries()) console.log("e", k, v);
for (const [k, v] of m) console.log("m", k, v);
console.log([...m.keys()].join(","), [...m.entries()], [...m]);
const set = new Set<string>(["x"]);
for (const [a, b] of set.entries()) console.log(a, b);

class Range {
  lo: number;
  hi: number;
  constructor(lo: number, hi: number) {
    this.lo = lo;
    this.hi = hi;
  }
  *[Symbol.iterator](): Generator<number> {
    for (let i = this.lo; i < this.hi; i++) yield i;
  }
  *ends(): Generator<number> {
    yield this.lo;
    yield this.hi;
  }
}
for (const x of new Range(1, 4)) console.log(x);
console.log([...new Range(0, 3)].join(","), [...new Range(5, 9).ends()].join(","));

function* pairOf(o: { lo: number }, n: number): Generator<number> {
  yield o.lo;
  yield n;
}
function wrap(): Generator<number> {
  return pairOf({ lo: 7 }, 8);
}
const g = wrap();
const first = g.next();
console.log(first.value, first.done, g.next().value, g.next().done);
const letters = function* (): Generator<string> {
  yield "p";
  yield "q";
};
for (const c of letters()) console.log(c);
