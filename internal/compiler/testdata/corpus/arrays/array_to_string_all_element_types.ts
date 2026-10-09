// @expect: 1,2 a,b 1,2 true,false
// @expect: 3,4 true 1.5,0.30000000000000004
// String(array) and toString() join every element type with commas.
console.log(String([1, 2]), String(["a", "b"]), String([1n, 2n]), String([true, false]));
const xs: bigint[] = [3n, 4n];
const bs: boolean[] = [true];
console.log(xs.toString(), bs.toString(), [1.5, 0.1 + 0.2].toString());
