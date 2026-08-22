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
