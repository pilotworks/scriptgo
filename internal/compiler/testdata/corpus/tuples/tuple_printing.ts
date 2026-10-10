// @expect: [ 'a', 1 ]
// @expect: [ [ 'a', 1 ], [ 'b', 2 ] ]
// @expect: { p: [ 'a', 1 ], n: 3 }
// @expect: Map(1) { 'k' => [ 1, 2 ] }
// @expect: [ [ 'x', 1 ], [ 'y', 2 ] ]
// @expect: [["a",1],["b",2]] {"p":["a",1],"n":3}
// @expect: { '0': 'a', '1': 1 }
// @expect: Map(2) { 'a' => 1, 'b' => 2 } 1 2
// Tuples print as arrays wherever they appear; an object literal with
// numeric keys stays an object, and Map entries keep their value types.
const pair: [string, number] = ["a", 1];
console.log(pair);
const list: [string, number][] = [
  ["a", 1],
  ["b", 2],
];
console.log(list);
const obj = { p: pair, n: 3 };
console.log(obj);
const m = new Map<string, [number, number]>([["k", [1, 2]]]);
console.log(m);
console.log(Object.entries({ x: 1, y: 2 }));
console.log(JSON.stringify(list), JSON.stringify(obj));
const numericKeys = { 0: "a", 1: 1 };
console.log(numericKeys);
const counts = new Map<string, number>([
  ["a", 1],
  ["b", 2],
]);
console.log(counts, counts.get("a"), counts.get("b"));
