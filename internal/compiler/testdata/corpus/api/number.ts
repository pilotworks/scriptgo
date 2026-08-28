// ScriptGo Corpus: Number Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: number.constants
// @api: Number.MAX_SAFE_INTEGER
// @api: Number.MIN_SAFE_INTEGER
// @expect: 9007199254740991
// @expect: -9007199254740991
// @expect: true
console.log(Number.MAX_SAFE_INTEGER); console.log(Number.MIN_SAFE_INTEGER); console.log(Number.EPSILON > 0);

// @api: number.valueOf
// @expect: 42
const numVal = 42;
console.log(numVal.valueOf());

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

// @api: number.isSafeInteger
// @expect: true
// @expect: false
console.log(Number.isSafeInteger(9007199254740991));
console.log(Number.isSafeInteger(9007199254740992));

// @api: Number.EPSILON
// @expect: true
console.log(Number.EPSILON > 0);

// @api: Number.MAX_VALUE
// @expect: true
console.log(Number.MAX_VALUE > 1e300);

// @api: Number.MIN_VALUE
// @expect: true
console.log(Number.MIN_VALUE > 0);

// @api: Number.NaN
// @expect: true
console.log(Number.isNaN(Number.NaN));

// @api: Number.POSITIVE_INFINITY
// @expect: true
console.log(Number.POSITIVE_INFINITY > 1e300);

// @api: Number.NEGATIVE_INFINITY
// @expect: true
console.log(Number.NEGATIVE_INFINITY < -1e300);

// @api: Number.toExponential
// @expect: 1.23e+2
console.log((123.456).toExponential(2));

// @api: Number.toPrecision
// @expect: 123
console.log((123.456).toPrecision(3));

// @api: Number.toLocaleString
// @expect: 123
console.log((123).toLocaleString());

// @api: Number.new
// @expect: 42
console.log(Number(42));

