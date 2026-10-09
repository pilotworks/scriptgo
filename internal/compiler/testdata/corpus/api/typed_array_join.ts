// @expect: 1,-2,3 1 - -2 - 3 1,-2,3
// @expect: 1|-5
// @expect: Int16Array(3) [ 1, -2, 3 ]
// TypedArray join/toString format elements with Number/BigInt toString.
const t = new Int16Array([1, -2, 3]);
console.log(t.join(), t.join(" - "), t.toString());
const b = new BigInt64Array([1n, -5n]);
console.log(b.join("|"));
console.log(t);
