// ScriptGo Corpus: Error Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: error.error
// @expect: fail
// @expect: Error
const e_error_error_0 = new Error("fail"); console.log(e_error_error_0.message); console.log(e_error_error_0.name);

// @api: error.range_error
// @expect: range fail
// @expect: RangeError
const e_error_range_error_1 = new RangeError("range fail"); console.log(e_error_range_error_1.message); console.log(e_error_range_error_1.name);

// @api: error.type_error
// @expect: type fail
// @expect: TypeError
const e_error_type_error_2 = new TypeError("type fail"); console.log(e_error_type_error_2.message); console.log(e_error_type_error_2.name);
