// @expect: 30
// @expect: 0
// @expect: 2147483647
// @expect: 1073741823
function bitwiseOperations(a: number, b: number): number {
    const andVal = a & b;
    const orVal = a | b;
    const xorVal = a ^ b;
    const notVal = ~a;
    const shlVal = a << 2;
    const shrVal = a >> 1;
    const zshrVal = b >>> 1;
    return (orVal ^ andVal) | (shlVal & 30);
}

console.log(bitwiseOperations(10, 20));

// Test 32-bit wrap around and unsigned right shift
const maxSigned = (1 << 30) - 1 + (1 << 30);
console.log(maxSigned ^ maxSigned);
console.log(maxSigned >>> 0);
console.log(maxSigned >>> 1);
