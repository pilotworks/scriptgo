const a: number = 0x12345678;
const mask: number = 0x000000FF;

console.log(a & mask);
console.log((a >> 8) & mask);
console.log((a >> 16) & mask);
console.log((a >> 24) & mask);

const negativeVal: number = -16;
console.log(negativeVal >> 2);
console.log(negativeVal >>> 2);

const combined: number = (0x0F | 0xF0) ^ 0xAA;
console.log(combined);

const inverted: number = ~0;
console.log(inverted);

const shiftWrap: number = 1 << 31;
console.log(shiftWrap);
console.log(shiftWrap >> 31);
console.log(shiftWrap >>> 31);
