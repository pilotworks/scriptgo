// @expect: 120
// @expect: 110
// @expect: 220
// @expect: 55
// @expect: 6
// @expect: 9
// @expect: 13
// @expect: 15
// @expect: 60
// @expect: 30
let a = 100;
a += 20;
console.log(a);
a -= 10;
console.log(a);
a *= 2;
console.log(a);
a /= 4;
console.log(a);
a %= 7;
console.log(a);

let b = 15;
b &= 9;
console.log(b);
b |= 4;
console.log(b);
b ^= 2;
console.log(b);
b <<= 2;
console.log(b);
b >>= 1;
console.log(b);
