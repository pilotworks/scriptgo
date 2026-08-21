const u8 = new Uint8Array(4);
u8[0] = 10;
u8[1] = 255;
u8[2] = 300; // truncated to 44 (300 & 0xff)
u8[3] = 42;

console.log(u8.length);
console.log(u8.byteLength);
console.log(u8.byteOffset);
console.log(u8.BYTES_PER_ELEMENT);
console.log(u8[0]);
console.log(u8[1]);
console.log(u8[2]);
console.log(u8[3]);
