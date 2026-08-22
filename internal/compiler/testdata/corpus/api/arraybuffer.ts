// ScriptGo Corpus: Arraybuffer Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: arraybuffer.byteLength
// @expect: 16
const buf_arraybuffer_byteLength_0 = new ArrayBuffer(16);
console.log(buf_arraybuffer_byteLength_0.byteLength);

// @api: arraybuffer.isView
// @expect: true
// @expect: false
const u8_arraybuffer_isView_1 = new Uint8Array([1, 2, 3]);
console.log(ArrayBuffer.isView(u8_arraybuffer_isView_1));
console.log(ArrayBuffer.isView("not a view"));

// @api: arraybuffer.slice
// @expect: 8
const buf_arraybuffer_slice_2 = new ArrayBuffer(16);
const sliced_arraybuffer_slice_2 = buf_arraybuffer_slice_2.slice(4, 12);
console.log(sliced_arraybuffer_slice_2.byteLength);
