let a = 12; // 0b1100
let b = 5;  // 0b0101

a &= b;
console.log(a); // 4 (0b0100)

a |= 2;
console.log(a); // 6 (0b0110)

a ^= 3;
console.log(a); // 5 (0b0101)

let c = 1;
c <<= 4;
console.log(c); // 16

c >>= 2;
console.log(c); // 4

let neg = -16;
neg >>>= 2;
console.log(neg); // 1073741820

// Logical assignment operators
let x: number = 0;
let y: number = 42;

x ||= y;
console.log(x); // 42

x &&= 100;
console.log(x); // 100

let n: number | null = null;
n ??= 50;
console.log(n); // 50

n ??= 999;
console.log(n); // 50
