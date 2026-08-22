// ScriptGo Corpus: Number Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: number.constants
// @expect: 9007199254740991
// @expect: -9007199254740991
// @expect: true
console.log(Number.MAX_SAFE_INTEGER); console.log(Number.MIN_SAFE_INTEGER); console.log(Number.EPSILON > 0);

// @api: number.isFinite
// @expect: true
// @expect: false
console.log(Number.isFinite(42)); console.log(Number.isFinite(Infinity));

// @api: number.isInteger
// @expect: true
// @expect: false
console.log(Number.isInteger(42)); console.log(Number.isInteger(4.2));

// @api: number.isNaN
// @expect: true
// @expect: false
console.log(Number.isNaN(NaN)); console.log(Number.isNaN(42));

// @api: number.parseFloat
// @expect: 3.14
// @expect: 2.718
console.log(Number.parseFloat("3.14")); console.log(parseFloat("2.718"));

// @api: number.parseInt
// @expect: 123
// @expect: 456
console.log(Number.parseInt("123")); console.log(parseInt("456"));

// @api: number.toFixed
// @expect: 3.14
const n_number_toFixed_6: number = 3.14159;
console.log(n_number_toFixed_6.toFixed(2));

// @api: number.toString
// @expect: 42
const n_number_toString_7: number = 42;
console.log(n_number_toString_7.toString());
