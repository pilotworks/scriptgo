// ScriptGo Corpus: Uint8array Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: uint8array.from
// @expect: 3
// @expect: 10
const arr_uint8array_from_0 = Uint8Array.from([10, 20, 30]);
console.log(arr_uint8array_from_0.length);
console.log(arr_uint8array_from_0[0]);

// @api: uint8array.constructor.typed-array-copy
// @expect: 7
// @expect: 42
const source_uint8array_copy = new Uint8Array(1);
source_uint8array_copy[0] = 7;
const copy_uint8array_copy = new Uint8Array(source_uint8array_copy);
copy_uint8array_copy[0] = 42;
console.log(source_uint8array_copy[0]);
console.log(copy_uint8array_copy[0]);
