const buf = new ArrayBuffer(16);
console.log(buf.byteLength);
console.log(ArrayBuffer.isView(buf));

const u8 = new Uint8Array(buf, 4, 8);
console.log(u8.length);
console.log(u8.byteLength);
console.log(u8.byteOffset);
console.log(ArrayBuffer.isView(u8));

u8[0] = 99;
const u8All = new Uint8Array(buf);
console.log(u8All[4]);
