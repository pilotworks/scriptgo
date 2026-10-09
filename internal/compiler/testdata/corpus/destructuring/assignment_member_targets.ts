// @expect: 5 6 hi 9 given 7
// Destructuring assignment stores into member targets (obj.prop, arr[i]),
// applies element defaults, and recurses into nested patterns.
const box = { a: 0, b: "" };
const list: number[] = [0, 0];
let n = 0;
let s = "";
let deep = 0;
[box.a, list[1]] = [5, 6];
const source: { x: string; y?: number } = { x: "hi" };
({ x: box.b, y: n = 9 } = source);
const words: string[] = ["given"];
[s = "dflt"] = words;
[[deep]] = [[7]];
console.log(box.a, list[1], box.b, n, s, deep);
