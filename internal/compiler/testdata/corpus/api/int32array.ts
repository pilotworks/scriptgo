// ScriptGo Corpus: Int32array Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: int32array.from
// @expect: 3
// @expect: 200
const arr_int32array_from_0 = Int32Array.from([100, 200, 300]);
console.log(arr_int32array_from_0.length);
console.log(arr_int32array_from_0[1]);
