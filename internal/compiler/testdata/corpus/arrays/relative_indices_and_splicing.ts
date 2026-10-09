// @expect: a,b b,c 2 1,2,3
// @expect: a,z,y,c a,b,q,c 9,3 a,b,c
// @expect: 0 8,8,8
// @expect: 2 2 2
// @expect: 3
// slice/toSpliced use relative indices (negative counts from the end, an
// omitted end means the length); toSpliced inserts its items; fill on an
// empty array is a no-op; for (;;) loops until break.
const letters = ["a", "b", "c"];
const nums = [1, 2, 3];
console.log(letters.slice(0, -1).join(), letters.slice(-2).join(), nums.slice(1, -1).join(), nums.slice(-5, 10).join());
console.log(letters.toSpliced(1, 1, "z", "y").join(), letters.toSpliced(-1, 0, "q").join(), nums.toSpliced(0, 2, 9).join(), letters.join());
const empty: number[] = [];
console.log(empty.fill(8).length, [0, 0, 0].fill(8).join());
const tuple: [number, string, boolean] = [1, "x", true];
console.log(tuple.slice(0, -1).length, tuple.slice(-2).length, tuple.slice(1).length);
let spins = 0;
for (;;) {
  spins++;
  if (spins === 3) break;
}
console.log(spins);
