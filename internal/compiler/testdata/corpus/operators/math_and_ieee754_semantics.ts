// @expect: NaN
// @expect: NaN
// @expect: -10
// @expect: 5
// @expect: NaN
// @expect: NaN
// @expect: -Infinity
// @expect: -Infinity
// @expect: -Infinity
// @expect: -Infinity
// @expect: -3.141592653589793
// @expect: -0
// @expect: -Infinity
// @expect: NaN
// @expect: 1
// @expect: NaN
// @expect: NaN
// @expect: -Infinity
// @expect: -Infinity
// @expect: -Infinity
// @expect: NaN
// 1. Math.min/max with NaN and Infinities
const nanVal: number = NaN;
const five: number = 5;
const negTen: number = -10;

console.log(Math.min(nanVal, five));
console.log(Math.max(nanVal, five));
console.log(Math.min(five, negTen));
console.log(Math.max(five, negTen));
console.log(Math.min(nanVal, Infinity));
console.log(Math.max(nanVal, -Infinity));

// 2. Signed zero (-0.0) propagation in Math functions
console.log(1 / Math.sign(-0));
console.log(1 / Math.sqrt(-0));
console.log(1 / Math.ceil(-0.5));
console.log(1 / Math.round(-0.5));
console.log(Math.atan2(-0, -0));
console.log(1 / Math.pow(-0, -3));
console.log(1 / Math.cbrt(-0));

// 3. ECMAScript Math.pow special semantics vs C99 pow
console.log(Math.pow(1, NaN));
console.log(Math.pow(NaN, 0));
console.log(Math.pow(-1, Infinity));
console.log(Math.pow(-1, -Infinity));

// 4. Math.expm1 and Math.log1p boundaries
console.log(1 / Math.expm1(-0));
console.log(1 / Math.log1p(-0));
console.log(Math.log1p(-1));
console.log(Math.log1p(-2));
