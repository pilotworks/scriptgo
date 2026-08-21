const b1: Buffer = Buffer.alloc(5);
console.log(b1.length);
console.log(b1[0]);

const b2: Buffer = Buffer.alloc(4, 42);
console.log(b2[0]);
console.log(b2[3]);

const b3: Buffer = Buffer.alloc(6, "abc");
console.log(b3[0]);
console.log(b3[1]);
console.log(b3[2]);
console.log(b3[3]);

const b4: Buffer = Buffer.allocUnsafe(3);
console.log(b4.length);
