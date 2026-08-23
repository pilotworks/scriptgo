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

// @api: BigInt.asIntN
// @expect: -2
// @expect: 2
const bi_asIntN = BigInt.asIntN(2, 6n);
console.log(bi_asIntN.toString());
const bi_asIntN_pos = BigInt.asIntN(4, 2n);
console.log(bi_asIntN_pos.toString());

// @api: BigInt.asUintN
// @expect: 2
// @expect: 14
const bi_asUintN = BigInt.asUintN(2, 6n);
console.log(bi_asUintN.toString());
const bi_asUintN_neg = BigInt.asUintN(4, -2n);
console.log(bi_asUintN_neg.toString());

// @api: BigInt.toLocaleString
// @expect: 987654321
const bi_locale = 987654321n;
console.log(bi_locale.toLocaleString());

// @api: BigInt.valueOf
// @expect: 987654321
console.log(bi_locale.valueOf().toString());
