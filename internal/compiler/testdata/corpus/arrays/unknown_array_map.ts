// @expect: number,string,boolean
// @expect: 1,2,3
// @expect: n1|sa|btrue
// Callbacks over an unknown[] receive each boxed element with its own type.
const values: unknown[] = [1, "a", true];
console.log(values.map((v) => typeof v).join(","));
console.log(values.map((_v, i) => i + 1).join(","));
console.log(values.map((v) => (typeof v).charAt(0) + String(v)).join("|"));
