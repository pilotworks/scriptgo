// @check.err: SG2005
// Native bigint is a signed 64-bit integer; wider literals are rejected.
const wide = 0xfedcba9876543210n;
console.log(wide);
