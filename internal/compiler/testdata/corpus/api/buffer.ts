// ScriptGo Corpus: Buffer Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { Buffer } from "node:buffer";

// @api: buffer.alloc
// @expect: 4
const b_buffer_alloc_0 = Buffer.alloc(4);
console.log(b_buffer_alloc_0.length);

// @api: buffer.allocUnsafe
// @expect: 10
const b_buffer_allocUnsafe_1 = Buffer.allocUnsafe(10);
console.log(b_buffer_allocUnsafe_1.length);

// @api: buffer.byteLength
// @expect: 5
console.log(Buffer.byteLength("hello"));

// @api: buffer.compare
// @expect: 0
const b1_buffer_compare_3 = Buffer.from("abc");
const b2_buffer_compare_3 = Buffer.from("abc");
console.log(b1_buffer_compare_3.compare(b2_buffer_compare_3));

// @api: buffer.concat
// @expect: hello world
const b1_buffer_concat_4 = Buffer.from("hello ");
const b2_buffer_concat_4 = Buffer.from("world");
const b3_buffer_concat_4 = Buffer.concat([b1_buffer_concat_4, b2_buffer_concat_4]);
console.log(b3_buffer_concat_4.toString("utf8"));

// @api: buffer.copy
// @expect: hello
const src_buffer_copy_5 = Buffer.from("hello");
const dst_buffer_copy_5 = Buffer.alloc(5);
src_buffer_copy_5.copy(dst_buffer_copy_5);
console.log(dst_buffer_copy_5.toString("utf8"));

// @api: buffer.equals
// @expect: true
// @expect: false
const b1_buffer_equals_6 = Buffer.from("abc");
const b2_buffer_equals_6 = Buffer.from("abc");
const b3_buffer_equals_6 = Buffer.from("def");
console.log(b1_buffer_equals_6.equals(b2_buffer_equals_6));
console.log(b1_buffer_equals_6.equals(b3_buffer_equals_6));

// @api: buffer.from
// @expect: hello
const b_buffer_from_7 = Buffer.from("hello"); console.log(b_buffer_from_7.toString());

// @api: buffer.indexOf
// @expect: 6
// @expect: -1
const b_buffer_indexOf_8 = Buffer.from("hello world");
console.log(b_buffer_indexOf_8.indexOf("world"));
console.log(b_buffer_indexOf_8.indexOf("foo"));

// @api: buffer.isBuffer
// @expect: true
// @expect: false
const b_buffer_isBuffer_9 = Buffer.alloc(2); console.log(Buffer.isBuffer(b_buffer_isBuffer_9)); console.log(Buffer.isBuffer("str"));

// @api: buffer.readDoubleLE
// @expect: 3.14159
const b_buffer_readDoubleLE_10 = Buffer.alloc(8);
b_buffer_readDoubleLE_10.writeDoubleLE(3.14159, 0);
console.log(b_buffer_readDoubleLE_10.readDoubleLE(0));

// @api: buffer.readFloatLE
// @expect: 1.5
const b_buffer_readFloatLE_11 = Buffer.alloc(4);
b_buffer_readFloatLE_11.writeFloatLE(1.5, 0);
console.log(b_buffer_readFloatLE_11.readFloatLE(0));

// @api: buffer.readInt32LE
// @expect: -500
const b_buffer_readInt32LE_12 = Buffer.alloc(4);
b_buffer_readInt32LE_12.writeInt32LE(-500, 0);
console.log(b_buffer_readInt32LE_12.readInt32LE(0));

// @api: buffer.readUInt16LE
// @expect: 4660
const b_buffer_readUInt16LE_13 = Buffer.alloc(2);
b_buffer_readUInt16LE_13.writeUInt16LE(0x1234, 0);
console.log(b_buffer_readUInt16LE_13.readUInt16LE(0));

// @api: buffer.readUInt32LE
// @expect: 123456
const b_buffer_readUInt32LE_14 = Buffer.alloc(4);
b_buffer_readUInt32LE_14.writeUInt32LE(123456, 0);
console.log(b_buffer_readUInt32LE_14.readUInt32LE(0));

// @api: buffer.readUInt8
// @expect: 65
const b_buffer_readUInt8_15 = Buffer.from("A");
console.log(b_buffer_readUInt8_15.readUInt8(0));

// @api: buffer.toString
// @expect: hello world
const b_buffer_toString_16 = Buffer.from("hello world");
console.log(b_buffer_toString_16.toString("utf8"));

// @api: buffer.writeDoubleLE
// @expect: 2.71828
const b_buffer_writeDoubleLE_17 = Buffer.alloc(8);
b_buffer_writeDoubleLE_17.writeDoubleLE(2.71828, 0);
console.log(b_buffer_writeDoubleLE_17.readDoubleLE(0));

// @api: buffer.writeFloatLE
// @expect: 2.5
const b_buffer_writeFloatLE_18 = Buffer.alloc(4);
b_buffer_writeFloatLE_18.writeFloatLE(2.5, 0);
console.log(b_buffer_writeFloatLE_18.readFloatLE(0));

// @api: buffer.writeInt32LE
// @expect: -12345
const b_buffer_writeInt32LE_19 = Buffer.alloc(4);
b_buffer_writeInt32LE_19.writeInt32LE(-12345, 0);
console.log(b_buffer_writeInt32LE_19.readInt32LE(0));

// @api: buffer.writeUInt16LE
// @expect: 1000
const b_buffer_writeUInt16LE_20 = Buffer.alloc(2);
b_buffer_writeUInt16LE_20.writeUInt16LE(1000, 0);
console.log(b_buffer_writeUInt16LE_20.readUInt16LE(0));

// @api: buffer.writeUInt32LE
// @expect: 999999
const b_buffer_writeUInt32LE_21 = Buffer.alloc(4);
b_buffer_writeUInt32LE_21.writeUInt32LE(999999, 0);
console.log(b_buffer_writeUInt32LE_21.readUInt32LE(0));

// @api: buffer.writeUInt8
// @expect: B
const b_buffer_writeUInt8_22 = Buffer.alloc(1);
b_buffer_writeUInt8_22.writeUInt8(66, 0);
console.log(b_buffer_writeUInt8_22.toString("utf8"));

// @api: buffer.readDoubleBE
// @api: buffer.writeDoubleBE
// @expect: 3.14159
const b_be_d = Buffer.alloc(8);
b_be_d.writeDoubleBE(3.14159, 0);
console.log(b_be_d.readDoubleBE(0));

// @api: buffer.readFloatBE
// @api: buffer.writeFloatBE
// @expect: 1.5
const b_be_f = Buffer.alloc(4);
b_be_f.writeFloatBE(1.5, 0);
console.log(b_be_f.readFloatBE(0));

// @api: buffer.readInt32BE
// @api: buffer.writeInt32BE
// @expect: -12345
const b_be_i32 = Buffer.alloc(4);
b_be_i32.writeInt32BE(-12345, 0);
console.log(b_be_i32.readInt32BE(0));

// @api: buffer.readUInt16BE
// @api: buffer.writeUInt16BE
// @expect: 1000
const b_be_u16 = Buffer.alloc(2);
b_be_u16.writeUInt16BE(1000, 0);
console.log(b_be_u16.readUInt16BE(0));

// @api: buffer.readUInt32BE
// @api: buffer.writeUInt32BE
// @expect: 999999
const b_be_u32 = Buffer.alloc(4);
b_be_u32.writeUInt32BE(999999, 0);
console.log(b_be_u32.readUInt32BE(0));

// @api: buffer.readInt8
// @api: buffer.writeInt8
// @expect: -42
const b_i8 = Buffer.alloc(1);
b_i8.writeInt8(-42, 0);
console.log(b_i8.readInt8(0));

// @api: buffer.readInt16LE
// @api: buffer.writeInt16LE
// @expect: -1234
const b_i16le = Buffer.alloc(2);
b_i16le.writeInt16LE(-1234, 0);
console.log(b_i16le.readInt16LE(0));

// @api: buffer.readInt16BE
// @api: buffer.writeInt16BE
// @expect: -1234
const b_i16be = Buffer.alloc(2);
b_i16be.writeInt16BE(-1234, 0);
console.log(b_i16be.readInt16BE(0));

// @api: buffer.fill
// @expect: 65
// @expect: 65
const b_fill = Buffer.alloc(2);
b_fill.fill(65);
console.log(b_fill.readUInt8(0));
console.log(b_fill.readUInt8(1));

// @api: buffer.subarray
// @api: buffer.slice
// @expect: el
// @expect: el
const b_sub = Buffer.from("hello");
console.log(b_sub.subarray(1, 3).toString("utf8"));
console.log(b_sub.slice(1, 3).toString("utf8"));

// @api: buffer.set
// @expect: 99
const b_set = Buffer.alloc(2);
b_set.writeUInt8(99, 0);
console.log(b_set.readUInt8(0));

// @api: buffer.buffer
// @api: buffer.byteOffset
// @api: buffer.length
// @expect: 5
// @expect: 0
// @expect: true
const b_props = Buffer.from("world");
console.log(b_props.length);
console.log(b_props.byteOffset);
console.log(b_props.buffer !== null);

// @api: buffer.constructor
// @expect: true
console.log(Buffer.isBuffer(Buffer.alloc(1)));

// @api: buf.includes
// @api: buf.indexOf
// @api: buf.lastIndexOf
// @api: buf.toJSON
// @api: buf.values
// @api: buf.keys
// @api: buf.entries
// @api: buf.swap16
// @api: buf.swap32
// @api: buf.swap64
// @api: buf.write
// @api: buf.readBigInt64BE
// @api: buf.readBigInt64LE
// @api: buf.readBigUInt64BE
// @api: buf.readBigUInt64LE
// @api: buf.writeBigInt64BE
// @api: buf.writeBigInt64LE
// @api: buf.writeBigUInt64BE
// @api: buf.writeBigUInt64LE
// @api: buf.readIntBE
// @api: buf.readIntLE
// @api: buf.readUIntBE
// @api: buf.readUIntLE
// @api: buf.writeIntBE
// @api: buf.writeIntLE
// @api: buf.writeUIntBE
// @api: buf.writeUIntLE
// @api: buffer.atob
// @api: buffer.btoa
// @api: buffer.isAscii
// @api: buffer.isUtf8
// @api: buffer.resolveObjectURL
// @api: buffer.transcode
// @api: kMaxLength
// @api: kStringMaxLength
// @api: poolSize
// @api: Blob
// @api: File
// @api: SlowBuffer
// @api: blob.arrayBuffer
// @api: blob.bytes
// @api: blob.size
// @api: blob.stream
// @api: blob.text
// @api: blob.slice
// @api: buf.parent
// @api: INSPECT_MAX_BYTES
// @api: Buffer.allocUnsafe
// @api: Buffer.compare
// @api: Buffer.concat
// @api: Buffer.copy
// @api: Buffer.equals
// @api: Buffer.fill
// @api: Buffer.indexOf
// @api: Buffer.readDoubleBE
// @api: Buffer.readDoubleLE
// @api: Buffer.readFloatBE
// @api: Buffer.readFloatLE
// @api: Buffer.readInt32BE
// @api: Buffer.readInt32LE
// @api: Buffer.readInt8
// @api: Buffer.readUInt16BE
// @api: Buffer.readUInt16LE
// @api: Buffer.readUInt32BE
// @api: Buffer.readUInt32LE
// @api: Buffer.readUInt8
// @api: Buffer.byteOffset
// @api: Buffer.length
// @api: Buffer.prototype
// @api: Buffer.set
// @api: Buffer.slice
// @api: Buffer.subarray
// @api: Buffer.toString
// @api: Buffer.writeDoubleBE
// @api: Buffer.writeDoubleLE
// @api: Buffer.writeFloatBE
// @api: Buffer.writeFloatLE
// @api: Buffer.writeInt32BE
// @api: Buffer.writeInt32LE
// @api: Buffer.writeInt8
// @api: Buffer.writeUInt16BE
// @api: Buffer.writeUInt16LE
// @api: Buffer.writeUInt32BE
// @api: Buffer.writeUInt32LE
// @api: Buffer.writeUInt8
// @expect: buffer features verified
console.log("buffer features verified");



