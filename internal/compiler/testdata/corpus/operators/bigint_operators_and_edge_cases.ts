// @expect: 9
// @expect: -3
// @expect: 6
// @expect: -1
// @expect: -7
// @expect: -3 -3 3
// @expect: -1 1 -1
// @expect: 1
// @expect: 1024
// @expect: -8
// @expect: -1 255
// 1. Bitwise operations on negative BigInts
console.log((~(-10n)).toString());
console.log((-10n >> 2n).toString());
console.log((-10n & 15n).toString());
console.log((-10n | 15n).toString());
console.log((-10n ^ 15n).toString());

// 2. Division truncation and modulo
console.log((-7n / 2n).toString(), (7n / -2n).toString(), (-7n / -2n).toString());
console.log((-7n % 2n).toString(), (7n % -2n).toString(), (-7n % -2n).toString());

// 3. Exponentiation operator with BigInt
console.log((0n ** 0n).toString());
console.log((2n ** 10n).toString());
console.log(((-2n) ** 3n).toString());

// 4. asIntN and asUintN
console.log(BigInt.asIntN(8, 255n).toString(), BigInt.asUintN(8, -1n).toString());
