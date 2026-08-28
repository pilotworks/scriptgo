// @expect: ~0 = -1
// @expect: ~(-1) = 0
// @expect: ~NaN = -1
// @expect: ~Infinity = -1
// @expect: ~(-Infinity) = -1
// @expect: ~3.7 = -4
// @expect: ~(-3.7) = 2
// @expect: 0 | 0 = 0
// @expect: 5.9 | 0 = 5
// @expect: -5.9 | 0 = -5
// @expect: NaN | 0 = 0
// @expect: Infinity | 0 = 0
// @expect: 2147483647 & 65535 = 65535
// @expect: 123 ^ 456 = 435
// @expect: 1 << 30 = 1073741824
// @expect: 1 << 31 = -2147483648
// @expect: -2147483648 >> 1 = -1073741824
// @expect: -1 >>> 0 = 4294967295
// @expect: (-2147483648 >>> 1) = 1073741824
function testBitwiseEdgeCases(): void {
    // Bitwise NOT
    console.log("~0 = " + (~0));
    console.log("~(-1) = " + (~(-1)));
    console.log("~NaN = " + (~NaN));
    console.log("~Infinity = " + (~Infinity));
    console.log("~(-Infinity) = " + (~(-Infinity)));
    console.log("~3.7 = " + (~3.7));
    console.log("~(-3.7) = " + (~(-3.7)));

    // Bitwise OR, AND, XOR
    console.log("0 | 0 = " + (0 | 0));
    console.log("5.9 | 0 = " + (5.9 | 0));
    console.log("-5.9 | 0 = " + (-5.9 | 0));
    console.log("NaN | 0 = " + (NaN | 0));
    console.log("Infinity | 0 = " + (Infinity | 0));
    console.log("2147483647 & 65535 = " + (2147483647 & 65535));
    console.log("123 ^ 456 = " + (123 ^ 456));

    // Shifts
    console.log("1 << 30 = " + (1 << 30));
    console.log("1 << 31 = " + (1 << 31));
    console.log("-2147483648 >> 1 = " + (-2147483648 >> 1));
    console.log("-1 >>> 0 = " + (-1 >>> 0));
    console.log("(-2147483648 >>> 1) = " + (-2147483648 >>> 1));
}

testBitwiseEdgeCases();
