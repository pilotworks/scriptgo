// @expect: Infinity
// @expect: -Infinity
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: true
// @expect: -1
// @expect: true
// @expect: true
// @expect: true
// @expect: 2
// @expect: 3
// @expect: 1.00
// @expect: 123.5
// @expect: 1.23e+4
// @expect: 11111111
// @expect: 377
// @expect: ff
// @expect: 73
// @expect: 31
// @expect: 15
// @expect: 255
// @expect: 255
// @expect: 42
// @expect: 0
// @expect: 0
// @expect: 0
// @expect: -2147483648
// @expect: 0
// @expect: 4294967295
// @expect: 2147483648
// 1. Signed zero and Object.is identity
const posZero: number = 0;
const negZero: number = -0;

console.log(1 / posZero);
console.log(1 / negZero);
console.log(posZero === negZero);
console.log(Object.is(posZero, negZero));
console.log(Object.is(posZero, 0));
console.log(Object.is(negZero, -0));

// 2. SameValueZero in Array operations
const numArr: number[] = [1, NaN, 2, -0];
console.log(numArr.includes(NaN));
console.log(numArr.indexOf(NaN));
console.log(numArr.includes(2));
console.log(numArr.includes(0));
console.log(numArr.includes(-0));

// 3. Number precision, formatting and radix
console.log((1.5).toFixed(0));
console.log((2.5).toFixed(0));
console.log((1.005).toFixed(2));
console.log((123.456).toFixed(1));
console.log((12345).toExponential(2));

const num255: number = 255;
console.log(num255.toString(2));
console.log(num255.toString(8));
console.log(num255.toString(16));
console.log(num255.toString(36));

// 4. parseInt with various radixes and hex prefixes
console.log(parseInt("0x1f", 16));
console.log(parseInt("1111", 2));
console.log(parseInt("377", 8));
console.log(parseInt("73", 36));
console.log(parseInt("   42px   "));

// 5. Bitwise OR coercion on non-finite floats and overflow
console.log(NaN | 0);
console.log(Infinity | 0);
console.log((-Infinity) | 0);
console.log(2147483648 | 0);
console.log(4294967296 | 0);
console.log((-1) >>> 0);
console.log((-2147483648) >>> 0);
