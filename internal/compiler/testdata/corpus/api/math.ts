// ScriptGo Corpus: Math Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: math.abs
// @expect: 42
// @expect: 42
console.log(Math.abs(-42)); console.log(Math.abs(42));

// @api: math.atan
// @expect: 0
console.log(Math.atan(0));

// @api: math.atan2
// @expect: 0
console.log(Math.atan2(0, 1));

// @api: math.ceil
// @expect: 5
// @expect: -4
console.log(Math.ceil(4.2)); console.log(Math.ceil(-4.8));

// @api: math.constants
// @expect: true
// @expect: true
// @expect: true
console.log(Math.PI > 3.14); console.log(Math.E > 2.71); console.log(Math.LN2 > 0.69);

// @api: math.cos
// @expect: 1
console.log(Math.cos(0));

// @api: math.exp
// @expect: 1
console.log(Math.exp(0));

// @api: math.floor
// @expect: 4
// @expect: -5
console.log(Math.floor(4.8)); console.log(Math.floor(-4.2));

// @api: math.hypot
// @expect: 5
console.log(Math.hypot(3, 4));

// @api: math.log
// @expect: 1
console.log(Math.log(Math.E));

// @api: math.log10
// @expect: 2
console.log(Math.log10(100));

// @api: math.log2
// @expect: 3
console.log(Math.log2(8));

// @api: math.max
// @expect: 20
// @expect: 5
console.log(Math.max(10, 20)); console.log(Math.max(-5, 5));

// @api: math.min
// @expect: 10
// @expect: -5
console.log(Math.min(10, 20)); console.log(Math.min(-5, 5));

// @api: math.pow
// @expect: 256
// @expect: 27
console.log(Math.pow(2, 8)); console.log(Math.pow(3, 3));

// @api: math.random
// @expect: true
const r_math_random_15 = Math.random(); console.log(r_math_random_15 >= 0 && r_math_random_15 < 1);

// @api: math.round
// @expect: 5
// @expect: 4
console.log(Math.round(4.5)); console.log(Math.round(4.4));

// @api: math.sign
// @expect: -1
// @expect: 1
// @expect: 0
console.log(Math.sign(-10)); console.log(Math.sign(10)); console.log(Math.sign(0));

// @api: math.sin
// @expect: 0
console.log(Math.sin(0));

// @api: math.sqrt
// @expect: 4
// @expect: 1.4142135623730951
console.log(Math.sqrt(16)); console.log(Math.sqrt(2));

// @api: math.tan
// @expect: 0
console.log(Math.tan(0));

// @api: math.trunc
// @expect: 4
// @expect: -4
console.log(Math.trunc(4.9)); console.log(Math.trunc(-4.9));
