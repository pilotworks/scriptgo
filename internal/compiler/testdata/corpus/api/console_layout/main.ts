// console.log lays out containers as util.inspect does: one line within
// breakLength 80, otherwise one entry per line; arrays of more than six
// entries in aligned columns; nesting past depth 2 as [Object] / [Array];
// at most 100 array elements; long multi-line strings split per line.
const nums: number[] = [];
for (let i = 0; i < 30; i++) nums.push(i * 7);
console.log(nums);
console.log([1, 2, 3, 4, 5, 6, 7]);
console.log(["apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew"]);
console.log({ a: { b: { c: { d: 1 } } } });
console.log({ a: { b: { c: 1 } } });
console.log([{ name: "Alice", age: 30, city: "Hanoi" }, { name: "Bob", age: 25, city: "Saigon" }, { name: "Carol", age: 41, city: "Danang" }]);
class Point { x: number; y: number; constructor(x: number, y: number) { this.x = x; this.y = y; } }
console.log(new Point(1, 2), [new Point(1, 2), new Point(3, 4), new Point(5, 6), new Point(7, 8), new Point(9, 10)]);
const m = new Map<string, { id: number; tags: string[] }>();
m.set("first", { id: 1, tags: ["x", "y", "z"] });
m.set("second", { id: 2, tags: ["long-tag-name", "another-long-tag"] });
console.log(m);
console.log(new Set(["alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota", "kappa"]));
console.log({ text: "line one of the text\nline two of the text\nline three of the text that is long" });
const big: number[] = [];
for (let i = 0; i < 120; i++) big.push(i);
console.log(big);
console.log([1.5, -2, 300, 4, 5, 6, 7, 8, 9, 10, 11]);
console.log("label", { alpha: "aaaaaaaaaaaa", beta: "bbbbbbbbbbbbbbbb", gamma: "cccccccccccccccc", delta: "ddddddddd" });
console.log([[1, 2, [3, 4, [5, 6]]], { k: [1, { z: 2 }] }]);
console.log([true, false, true, false, true, false, true, false]);
const u: unknown = { name: "Alice", tags: ["admin", "editor", "viewer"], profile: { city: "Hanoi" } };
console.log("user", u);
const g = /(?<year>\d{4})-(?<month>\d{2})-(?<day>\d{2})/.exec("on 2024-05-06")!;
console.log(g);
