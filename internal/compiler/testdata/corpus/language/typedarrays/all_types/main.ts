// Int8Array
const i8 = new Int8Array(3);
i8[0] = 127;
i8[1] = -128;
i8[2] = 200;
console.log(i8.BYTES_PER_ELEMENT);
console.log(i8.byteLength);
console.log(i8[0]);
console.log(i8[1]);
console.log(i8[2]);

// Uint8ClampedArray
const u8c = new Uint8ClampedArray(4);
u8c[0] = -10;
u8c[1] = 120.4;
u8c[2] = 120.6;
u8c[3] = 300;
console.log(u8c.BYTES_PER_ELEMENT);
console.log(u8c[0]);
console.log(u8c[1]);
console.log(u8c[2]);
console.log(u8c[3]);

// Int16Array & Uint16Array
const i16 = new Int16Array(2);
i16[0] = 32767;
i16[1] = -32768;
console.log(i16.BYTES_PER_ELEMENT);
console.log(i16.byteLength);
console.log(i16[0]);
console.log(i16[1]);

const u16 = new Uint16Array(2);
u16[0] = 65535;
u16[1] = 100;
console.log(u16.BYTES_PER_ELEMENT);
console.log(u16.byteLength);
console.log(u16[0]);
console.log(u16[1]);

// Uint32Array & Float32Array
const u32 = new Uint32Array(2);
u32[0] = 4294967295;
u32[1] = 500;
console.log(u32.BYTES_PER_ELEMENT);
console.log(u32.byteLength);
console.log(u32[0]);
console.log(u32[1]);

const f32 = new Float32Array(2);
f32[0] = 1.5;
f32[1] = -2.5;
console.log(f32.BYTES_PER_ELEMENT);
console.log(f32.byteLength);
console.log(f32[0]);
console.log(f32[1]);

// BigInt64Array & BigUint64Array
const bi64 = new BigInt64Array(2);
bi64[0] = 9007199254740993n;
bi64[1] = -9007199254740993n;
console.log(bi64.BYTES_PER_ELEMENT);
console.log(bi64.byteLength);
console.log(bi64[0]);
console.log(bi64[1]);

const bu64 = new BigUint64Array(2);
bu64[0] = 1844674407370955161n;
bu64[1] = 42n;
console.log(bu64.BYTES_PER_ELEMENT);
console.log(bu64.byteLength);
console.log(bu64[0]);
console.log(bu64[1]);

console.log(ArrayBuffer.isView(i8));
console.log(ArrayBuffer.isView(u8c));
console.log(ArrayBuffer.isView(bi64));
console.log(ArrayBuffer.isView("not a view"));
