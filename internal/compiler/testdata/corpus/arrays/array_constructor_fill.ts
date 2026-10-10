// new Array(n).fill(v) and Array(n).fill(v) produce an array of v's type.
// @expect: 0,2,4,6,8
// @expect: x-x-x 3
// @expect: y+y
// @expect: [[1,1],[1,1]]
const n = 5;
const a: number[] = new Array(n).fill(0);
for (let i = 0; i < n; i++) { a[i] = i * 2; }
console.log(a.join(","));
const b: string[] = new Array<string>(3).fill("x");
console.log(b.join("-"), b.length);
const c: string[] = Array(2).fill("y");
console.log(c.join("+"));
const grid: number[][] = new Array(2).fill(0).map(() => new Array<number>(2).fill(1));
console.log(JSON.stringify(grid));
