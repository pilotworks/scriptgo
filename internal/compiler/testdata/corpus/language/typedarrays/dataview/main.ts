const buf = new ArrayBuffer(16);
const dv = new DataView(buf, 0, 16);

console.log(dv.byteLength);
console.log(dv.byteOffset);

dv.setInt8(0, -12);
dv.setUint8(1, 250);
console.log(dv.getInt8(0));
console.log(dv.getUint8(1));

// 16-bit
dv.setInt16(2, 4660); // Big-Endian by default (0x1234)
console.log(dv.getInt16(2)); // 4660
console.log(dv.getInt16(2, true)); // 13330 (0x3412)

dv.setUint16(4, 22136, true); // Little-Endian write (0x5678)
console.log(dv.getUint16(4, true)); // 22136
console.log(dv.getUint16(4)); // 30806 (0x7856)

// 32-bit
dv.setInt32(6, -100000);
console.log(dv.getInt32(6));

// float
dv.setFloat32(0, 1.5, true);
console.log(dv.getFloat32(0, true));

dv.setFloat64(0, 3.141592653589793);
console.log(dv.getFloat64(0));

// BigInt
dv.setBigInt64(8, 9007199254740993n, true);
console.log(dv.getBigInt64(8, true));

console.log(ArrayBuffer.isView(dv));
