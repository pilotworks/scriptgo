// @expect: 1073741823
// @expect: -1
// @expect: 2147483648
// @expect: 1
// @expect: 40
// @expect: -10
const neg = -4;
console.log(neg >>> 2);
console.log(neg >> 2);

const shifted = (1 << 31) >>> 0;
console.log(shifted);

const bitwiseAnd = 5 & 3;
console.log(bitwiseAnd);

const bitwiseOr = 32 | 8;
console.log(bitwiseOr);

const bitwiseXor = ~9;
console.log(bitwiseXor);
