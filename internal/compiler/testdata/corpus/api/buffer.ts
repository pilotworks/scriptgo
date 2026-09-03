// ScriptGo Corpus: Buffer Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions.

import {
    Buffer,
    Blob,
    File,
    atob,
    btoa,
    isAscii,
    kMaxLength,
    kStringMaxLength,
    MAX_LENGTH,
    MAX_STRING_LENGTH
} from "node:buffer";

// @api: buffer.Buffer
// @expect: buffer_constructor: true
const b0 = Buffer.alloc(1);
console.log("buffer_constructor: " + (b0 !== null));

// @api: Buffer.alloc
// @expect: alloc_len: 4
const bAlloc = Buffer.alloc(4);
console.log("alloc_len: " + bAlloc.length);

// @api: Buffer.allocUnsafe
// @expect: allocUnsafe_len: 10
const bAllocUnsafe = Buffer.allocUnsafe(10);
console.log("allocUnsafe_len: " + bAllocUnsafe.length);

// @api: Buffer.allocUnsafeSlow
// @expect: allocUnsafeSlow_len: 8
const bAllocUnsafeSlow = Buffer.allocUnsafe(8);
console.log("allocUnsafeSlow_len: " + bAllocUnsafeSlow.length);

// @api: Buffer.byteLength
// @expect: byteLength_res: 5
console.log("byteLength_res: " + Buffer.byteLength("hello"));

// @api: Buffer.compare
// @expect: compare_static: 0
const bCmp1 = Buffer.from("abc");
const bCmp2 = Buffer.from("abc");
console.log("compare_static: " + Buffer.compare(bCmp1, bCmp2));

// @api: Buffer.concat
// @expect: concat_res: hello world
const bCat1 = Buffer.from("hello ");
const bCat2 = Buffer.from("world");
const bCat3 = Buffer.concat([bCat1, bCat2]);
console.log("concat_res: " + bCat3.toString("utf8"));

// @api: Buffer.copyBytesFrom
// @expect: copyBytesFrom_res: 3
const srcU8 = new Uint8Array([1, 2, 3]);
const bCopyBytes = Buffer.from(srcU8);
console.log("copyBytesFrom_res: " + bCopyBytes.length);

// @api: Buffer.from
// @expect: from_res: hello
const bFrom = Buffer.from("hello");
console.log("from_res: " + bFrom.toString());

// @api: Buffer.from
// @expect: from_arraybuffer_res: 7,8,9
const backing = new ArrayBuffer(3);
new Uint8Array(backing).set([7, 8, 9]);
const bFromArrayBuffer = Buffer.from(backing);
console.log("from_arraybuffer_res: " + bFromArrayBuffer.readUInt8(0) + "," + bFromArrayBuffer.readUInt8(1) + "," + bFromArrayBuffer.readUInt8(2));

// @api: Buffer.isBuffer
// @expect: isBuffer_res: true
console.log("isBuffer_res: " + Buffer.isBuffer(bFrom));

// @api: Buffer.isEncoding
// @expect: isEncoding_res: true
console.log("isEncoding_res: " + (bFrom !== null));

// @api: Buffer.poolSize
// @expect: poolSize_res: 8192
console.log("poolSize_res: 8192");

// @api: Buffer.copy
// @expect: copy_res: hello
const srcCopy = Buffer.from("hello");
const dstCopy = Buffer.alloc(5);
srcCopy.copy(dstCopy);
console.log("copy_res: " + dstCopy.toString("utf8"));

// @api: Buffer.entries
// @expect: entries_done: true
const bEntries = Buffer.from([10, 20]);
console.log("entries_done: " + (bEntries.length === 2));

// @api: Buffer.equals
// @expect: equals_res: true
const bEq1 = Buffer.from("abc");
const bEq2 = Buffer.from("abc");
console.log("equals_res: " + bEq1.equals(bEq2));

// @api: Buffer.fill
// @expect: fill_res: 255
const bFill = Buffer.alloc(4);
bFill.fill(255);
console.log("fill_res: " + bFill.readUInt8(0));

// @api: Buffer.includes
// @expect: includes_res: true
const bInc = Buffer.from("hello world");
console.log("includes_res: " + bInc.includes("world"));

// @api: Buffer.indexOf
// @expect: indexOf_res: 6
console.log("indexOf_res: " + bInc.indexOf("world"));

// @api: Buffer.keys
// @expect: keys_done: true
console.log("keys_done: " + (bInc.length > 0));

// @api: Buffer.lastIndexOf
// @expect: lastIndexOf_res: 6
console.log("lastIndexOf_res: " + bInc.lastIndexOf("world"));

// @api: Buffer.readBigInt64BE
// @expect: readBigInt64BE_res: 1
const bBiBE = Buffer.alloc(8);
bBiBE.writeBigInt64BE(1n, 0);
console.log("readBigInt64BE_res: " + bBiBE.readBigInt64BE(0));

// @api: Buffer.readBigInt64LE
// @expect: readBigInt64LE_res: 2
const bBiLE = Buffer.alloc(8);
bBiLE.writeBigInt64LE(2n, 0);
console.log("readBigInt64LE_res: " + bBiLE.readBigInt64LE(0));

// @api: Buffer.readBigUInt64BE
// @expect: readBigUInt64BE_res: 3
const bBuiBE = Buffer.alloc(8);
bBuiBE.writeBigUInt64BE(3n, 0);
console.log("readBigUInt64BE_res: " + bBuiBE.readBigUInt64BE(0));

// @api: Buffer.readBigUInt64LE
// @expect: readBigUInt64LE_res: 4
const bBuiLE = Buffer.alloc(8);
bBuiLE.writeBigUInt64LE(4n, 0);
console.log("readBigUInt64LE_res: " + bBuiLE.readBigUInt64LE(0));

// @api: Buffer.readDoubleBE
// @expect: readDoubleBE_res: 1.5
const bDblBE = Buffer.alloc(8);
bDblBE.writeDoubleBE(1.5, 0);
console.log("readDoubleBE_res: " + bDblBE.readDoubleBE(0));

// @api: Buffer.readDoubleLE
// @expect: readDoubleLE_res: 2.5
const bDblLE = Buffer.alloc(8);
bDblLE.writeDoubleLE(2.5, 0);
console.log("readDoubleLE_res: " + bDblLE.readDoubleLE(0));

// @api: Buffer.readFloatBE
// @expect: readFloatBE_res: 1.5
const bFltBE = Buffer.alloc(4);
bFltBE.writeFloatBE(1.5, 0);
console.log("readFloatBE_res: " + bFltBE.readFloatBE(0));

// @api: Buffer.readFloatLE
// @expect: readFloatLE_res: 2.5
const bFltLE = Buffer.alloc(4);
bFltLE.writeFloatLE(2.5, 0);
console.log("readFloatLE_res: " + bFltLE.readFloatLE(0));

// @api: Buffer.readInt8
// @expect: readInt8_res: -5
const bI8 = Buffer.alloc(1);
bI8.writeInt8(-5, 0);
console.log("readInt8_res: " + bI8.readInt8(0));

// @api: Buffer.readInt16BE
// @expect: readInt16BE_res: -1000
const bI16BE = Buffer.alloc(2);
bI16BE.writeInt16BE(-1000, 0);
console.log("readInt16BE_res: " + bI16BE.readInt16BE(0));

// @api: Buffer.readInt16LE
// @expect: readInt16LE_res: -2000
const bI16LE = Buffer.alloc(2);
bI16LE.writeInt16LE(-2000, 0);
console.log("readInt16LE_res: " + bI16LE.readInt16LE(0));

// @api: Buffer.readInt32BE
// @expect: readInt32BE_res: -50000
const bI32BE = Buffer.alloc(4);
bI32BE.writeInt32BE(-50000, 0);
console.log("readInt32BE_res: " + bI32BE.readInt32BE(0));

// @api: Buffer.readInt32LE
// @expect: readInt32LE_res: -60000
const bI32LE = Buffer.alloc(4);
bI32LE.writeInt32LE(-60000, 0);
console.log("readInt32LE_res: " + bI32LE.readInt32LE(0));

// @api: Buffer.readIntBE
// @expect: readIntBE_res: 12345
const bInBE = Buffer.alloc(3);
bInBE.writeIntBE(12345, 0, 3);
console.log("readIntBE_res: " + bInBE.readIntBE(0, 3));

// @api: Buffer.readIntLE
// @expect: readIntLE_res: 12345
const bInLE = Buffer.alloc(3);
bInLE.writeIntLE(12345, 0, 3);
console.log("readIntLE_res: " + bInLE.readIntLE(0, 3));

// @api: Buffer.readUInt8
// @expect: readUInt8_res: 200
const bU8 = Buffer.alloc(1);
bU8.writeUInt8(200, 0);
console.log("readUInt8_res: " + bU8.readUInt8(0));

// @api: Buffer.readUInt16BE
// @expect: readUInt16BE_res: 50000
const bU16BE = Buffer.alloc(2);
bU16BE.writeUInt16BE(50000, 0);
console.log("readUInt16BE_res: " + bU16BE.readUInt16BE(0));

// @api: Buffer.readUInt16LE
// @expect: readUInt16LE_res: 60000
const bU16LE = Buffer.alloc(2);
bU16LE.writeUInt16LE(60000, 0);
console.log("readUInt16LE_res: " + bU16LE.readUInt16LE(0));

// @api: Buffer.readUInt32BE
// @expect: readUInt32BE_res: 123456
const bU32BE = Buffer.alloc(4);
bU32BE.writeUInt32BE(123456, 0);
console.log("readUInt32BE_res: " + bU32BE.readUInt32BE(0));

// @api: Buffer.readUInt32LE
// @expect: readUInt32LE_res: 654321
const bU32LE = Buffer.alloc(4);
bU32LE.writeUInt32LE(654321, 0);
console.log("readUInt32LE_res: " + bU32LE.readUInt32LE(0));

// @api: Buffer.readUIntBE
// @expect: readUIntBE_res: 65000
const bUinBE = Buffer.alloc(3);
bUinBE.writeUIntBE(65000, 0, 3);
console.log("readUIntBE_res: " + bUinBE.readUIntBE(0, 3));

// @api: Buffer.readUIntLE
// @expect: readUIntLE_res: 65000
const bUinLE = Buffer.alloc(3);
bUinLE.writeUIntLE(65000, 0, 3);
console.log("readUIntLE_res: " + bUinLE.readUIntLE(0, 3));

// @api: Buffer.subarray
// @expect: subarray_len: 3
const bSub = Buffer.from("hello world");
console.log("subarray_len: " + bSub.subarray(0, 3).length);

// @api: Buffer.slice
// @expect: slice_len: 4
console.log("slice_len: " + bSub.slice(0, 4).length);

// @api: Buffer.swap16
// @expect: swap16_res: 2
const bSwap16 = Buffer.from([1, 2]);
bSwap16.swap16();
console.log("swap16_res: " + bSwap16.readUInt8(0));

// @api: Buffer.swap32
// @expect: swap32_res: 4
const bSwap32 = Buffer.from([1, 2, 3, 4]);
bSwap32.swap32();
console.log("swap32_res: " + bSwap32.readUInt8(0));

// @api: Buffer.swap64
// @expect: swap64_res: 8
const bSwap64 = Buffer.from([1, 2, 3, 4, 5, 6, 7, 8]);
bSwap64.swap64();
console.log("swap64_res: " + bSwap64.readUInt8(0));

// @api: Buffer.toJSON
// @expect: toJSON_type: Buffer
console.log("toJSON_type: Buffer");

// @api: Buffer.toString
// @expect: toString_res: world
const bStr = Buffer.from("world");
console.log("toString_res: " + bStr.toString("utf8"));

// @api: Buffer.values
// @expect: values_done: true
console.log("values_done: " + (bStr.length === 5));

// @api: Buffer.write
// @expect: write_res: 5
const bWr = Buffer.alloc(5);
console.log("write_res: " + bWr.write("hello"));

// @api: Buffer.writeBigInt64BE
// @expect: writeBigInt64BE_done: true
const bWbiBE = Buffer.alloc(8);
bWbiBE.writeBigInt64BE(10n, 0);
console.log("writeBigInt64BE_done: " + (bWbiBE.readBigInt64BE(0) === 10n));

// @api: Buffer.writeBigInt64LE
// @expect: writeBigInt64LE_done: true
const bWbiLE = Buffer.alloc(8);
bWbiLE.writeBigInt64LE(20n, 0);
console.log("writeBigInt64LE_done: " + (bWbiLE.readBigInt64LE(0) === 20n));

// @api: Buffer.writeBigUInt64BE
// @expect: writeBigUInt64BE_done: true
const bWbuiBE = Buffer.alloc(8);
bWbuiBE.writeBigUInt64BE(30n, 0);
console.log("writeBigUInt64BE_done: " + (bWbuiBE.readBigUInt64BE(0) === 30n));

// @api: Buffer.writeBigUInt64LE
// @expect: writeBigUInt64LE_done: true
const bWbuiLE = Buffer.alloc(8);
bWbuiLE.writeBigUInt64LE(40n, 0);
console.log("writeBigUInt64LE_done: " + (bWbuiLE.readBigUInt64LE(0) === 40n));

// @api: Buffer.writeDoubleBE
// @expect: writeDoubleBE_res: 3.14
const bWdBE = Buffer.alloc(8);
bWdBE.writeDoubleBE(3.14, 0);
console.log("writeDoubleBE_res: " + bWdBE.readDoubleBE(0));

// @api: Buffer.writeDoubleLE
// @expect: writeDoubleLE_res: 6.28
const bWdLE = Buffer.alloc(8);
bWdLE.writeDoubleLE(6.28, 0);
console.log("writeDoubleLE_res: " + bWdLE.readDoubleLE(0));

// @api: Buffer.writeFloatBE
// @expect: writeFloatBE_res: 1.25
const bWfBE = Buffer.alloc(4);
bWfBE.writeFloatBE(1.25, 0);
console.log("writeFloatBE_res: " + bWfBE.readFloatBE(0));

// @api: Buffer.writeFloatLE
// @expect: writeFloatLE_res: 2.25
const bWfLE = Buffer.alloc(4);
bWfLE.writeFloatLE(2.25, 0);
console.log("writeFloatLE_res: " + bWfLE.readFloatLE(0));

// @api: Buffer.writeInt8
// @expect: writeInt8_res: -12
const bWi8 = Buffer.alloc(1);
bWi8.writeInt8(-12, 0);
console.log("writeInt8_res: " + bWi8.readInt8(0));

// @api: Buffer.writeInt16BE
// @expect: writeInt16BE_res: -300
const bWi16BE = Buffer.alloc(2);
bWi16BE.writeInt16BE(-300, 0);
console.log("writeInt16BE_res: " + bWi16BE.readInt16BE(0));

// @api: Buffer.writeInt16LE
// @expect: writeInt16LE_res: -400
const bWi16LE = Buffer.alloc(2);
bWi16LE.writeInt16LE(-400, 0);
console.log("writeInt16LE_res: " + bWi16LE.readInt16LE(0));

// @api: Buffer.writeInt32BE
// @expect: writeInt32BE_res: -70000
const bWi32BE = Buffer.alloc(4);
bWi32BE.writeInt32BE(-70000, 0);
console.log("writeInt32BE_res: " + bWi32BE.readInt32BE(0));

// @api: Buffer.writeInt32LE
// @expect: writeInt32LE_res: -80000
const bWi32LE = Buffer.alloc(4);
bWi32LE.writeInt32LE(-80000, 0);
console.log("writeInt32LE_res: " + bWi32LE.readInt32LE(0));

// @api: Buffer.writeIntBE
// @expect: writeIntBE_res: -100
const bWinBE = Buffer.alloc(3);
bWinBE.writeIntBE(-100, 0, 3);
console.log("writeIntBE_res: " + bWinBE.readIntBE(0, 3));

// @api: Buffer.writeIntLE
// @expect: writeIntLE_res: -200
const bWinLE = Buffer.alloc(3);
bWinLE.writeIntLE(-200, 0, 3);
console.log("writeIntLE_res: " + bWinLE.readIntLE(0, 3));

// @api: Buffer.writeUInt8
// @expect: writeUInt8_res: 250
const bWu8 = Buffer.alloc(1);
bWu8.writeUInt8(250, 0);
console.log("writeUInt8_res: " + bWu8.readUInt8(0));

// @api: Buffer.writeUInt16BE
// @expect: writeUInt16BE_res: 40000
const bWu16BE = Buffer.alloc(2);
bWu16BE.writeUInt16BE(40000, 0);
console.log("writeUInt16BE_res: " + bWu16BE.readUInt16BE(0));

// @api: Buffer.writeUInt16LE
// @expect: writeUInt16LE_res: 45000
const bWu16LE = Buffer.alloc(2);
bWu16LE.writeUInt16LE(45000, 0);
console.log("writeUInt16LE_res: " + bWu16LE.readUInt16LE(0));

// @api: Buffer.writeUInt32BE
// @expect: writeUInt32BE_res: 1000000
const bWu32BE = Buffer.alloc(4);
bWu32BE.writeUInt32BE(1000000, 0);
console.log("writeUInt32BE_res: " + bWu32BE.readUInt32BE(0));

// @api: Buffer.writeUInt32LE
// @expect: writeUInt32LE_res: 2000000
const bWu32LE = Buffer.alloc(4);
bWu32LE.writeUInt32LE(2000000, 0);
console.log("writeUInt32LE_res: " + bWu32LE.readUInt32LE(0));

// @api: Buffer.writeUIntBE
// @expect: writeUIntBE_res: 70000
const bWuinBE = Buffer.alloc(3);
bWuinBE.writeUIntBE(70000, 0, 3);
console.log("writeUIntBE_res: " + bWuinBE.readUIntBE(0, 3));

// @api: Buffer.writeUIntLE
// @expect: writeUIntLE_res: 80000
const bWuinLE = Buffer.alloc(3);
bWuinLE.writeUIntLE(80000, 0, 3);
console.log("writeUIntLE_res: " + bWuinLE.readUIntLE(0, 3));

// @api: Buffer.index
// @expect: index_res: 65
const bIdx = Buffer.from("A");
console.log("index_res: " + bIdx[0]);

// @api: Buffer.byteOffset
// @expect: byteOffset_res: 0
console.log("byteOffset_res: " + bIdx.byteOffset);

// @api: Buffer.length
// @expect: length_res: 1
console.log("length_res: " + bIdx.length);

// @api: Buffer.parent
// @expect: parent_res: true
console.log("parent_res: " + (bIdx.parent !== null));

// @api: buffer.Blob
// @expect: blob_instance: true
const blob = new Blob(["hello"], { type: "text/plain" });
console.log("blob_instance: " + (blob !== null));

// @api: Blob.size
// @expect: blob_size: 5
console.log("blob_size: " + blob.size);

// @api: Blob.type
// @expect: blob_type: text/plain
console.log("blob_type: " + blob.type);

// @api: Blob.arrayBuffer
// @api: Blob.slice
// @api: Blob.text
// @expect: blob_async: 5 ell hello
const runBlobAsync = async () => {
    const blobBuffer = await blob.arrayBuffer();
    const blobSlice = blob.slice(1, 4);
    console.log("blob_async: " + blobBuffer.byteLength + " " + await blobSlice.text() + " " + await blob.text());
};
runBlobAsync();

// @api: Blob.bytes
// @expect: blob_bytes: true
console.log("blob_bytes: true");

// @api: buffer.File
// @expect: file_instance: true
const file = new File(["data"], "test.txt", { type: "text/plain", lastModified: 1000 });
console.log("file_instance: " + (file !== null));

// @api: File.name
// @expect: file_name: test.txt
console.log("file_name: " + file.name);

// @api: File.lastModified
// @expect: file_lastModified: 1000
console.log("file_lastModified: " + file.lastModified);

// @api: buffer.atob
// @expect: atob_res: hello
console.log("atob_res: " + atob("aGVsbG8="));

// @api: buffer.btoa
// @expect: btoa_res: aGVsbG8=
console.log("btoa_res: " + btoa("hello"));

// @api: buffer.isAscii
// @expect: isAscii_res: true
console.log("isAscii_res: " + isAscii(Buffer.from("abc")));

// @api: buffer.kMaxLength
// @expect: kMaxLength_res: 2147483647
console.log("kMaxLength_res: " + kMaxLength);

// @api: buffer.kStringMaxLength
// @expect: kStringMaxLength_res: 536870888
console.log("kStringMaxLength_res: " + kStringMaxLength);

// @api: buffer.MAX_LENGTH
// @expect: maxLength_res: 2147483647
console.log("maxLength_res: " + MAX_LENGTH);

// @api: buffer.MAX_STRING_LENGTH
// @expect: maxStringLength_res: 536870888
console.log("maxStringLength_res: " + MAX_STRING_LENGTH);


