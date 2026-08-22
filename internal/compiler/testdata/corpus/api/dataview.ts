// ScriptGo Corpus: Dataview Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: dataview.getFloat64
// @expect: 3.14
const buf_dataview_getFloat64_0 = new ArrayBuffer(8);
const dv_dataview_getFloat64_0 = new DataView(buf_dataview_getFloat64_0);
dv_dataview_getFloat64_0.setFloat64(0, 3.14);
console.log(dv_dataview_getFloat64_0.getFloat64(0));

// @api: dataview.getInt32
// @expect: 100000
const buf_dataview_getInt32_1 = new ArrayBuffer(4);
const dv_dataview_getInt32_1 = new DataView(buf_dataview_getInt32_1);
dv_dataview_getInt32_1.setInt32(0, 100000);
console.log(dv_dataview_getInt32_1.getInt32(0));

// @api: dataview.getInt8
// @expect: 42
const buf_dataview_getInt8_2 = new ArrayBuffer(4);
const dv_dataview_getInt8_2 = new DataView(buf_dataview_getInt8_2);
dv_dataview_getInt8_2.setInt8(0, 42);
console.log(dv_dataview_getInt8_2.getInt8(0));
