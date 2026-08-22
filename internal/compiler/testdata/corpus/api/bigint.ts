// ScriptGo Corpus: Bigint Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: bigint.constructor
// @expect: bigint
// @expect: bigint
const b1_bigint_constructor_0 = BigInt(42); const b2_bigint_constructor_0 = BigInt("100"); console.log(typeof b1_bigint_constructor_0); console.log(typeof b2_bigint_constructor_0);

// @api: bigint.toString
// @expect: 123456789012345
const b_bigint_toString_1 = 123456789012345n;
console.log(b_bigint_toString_1.toString());
