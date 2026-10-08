// @expect: -Infinity false true
// @expect: true -Infinity
// @expect: true true
// @expect: false true
// @expect: true -Infinity
// Signed zero must survive constant folding, and Object.is must compare
// mixed boxed/unboxed operands with SameValue semantics.
const z = -0;
console.log(1 / z, Object.is(z, 0), Object.is(z, -0));
const product = -1 * 0;
console.log(Object.is(product, -0), 1 / product);
const sum = -0 + -0;
console.log(Object.is(sum, -0), Object.is(-0 + 0, 0));
const boxedZero: unknown = 0;
const boxedNegative: unknown = -0;
console.log(Object.is(boxedZero, boxedNegative), Object.is(boxedNegative, -0));
console.log(Math.sign(-0) === 0, 1 / Math.sign(-0));
