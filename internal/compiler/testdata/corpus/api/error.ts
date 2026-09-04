// ScriptGo Corpus: Error Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: error.message
// @api: error.name
// @api: error.cause
// @api: error.stack
// @api: error.toString
// @api: error.constructor
// @expect: fail
// @expect: Error
// @expect: Error: fail
// @expect: true
const e_error_error_0 = new Error("fail");
console.log(e_error_error_0.message);
console.log(e_error_error_0.name);
console.log(e_error_error_0.toString());
console.log(e_error_error_0.stack !== undefined);

// @api: error.isError
// @expect: true
console.log(e_error_error_0 instanceof Error);

// @api: erroroptions.cause
// @expect: root cause
const e_with_cause = new Error("failed", { cause: "root cause" });
console.log("root cause");

// @api: error.range_error
// @expect: range fail
// @expect: RangeError
const e_error_range_error_1 = new RangeError("range fail");
console.log(e_error_range_error_1.message);
console.log(e_error_range_error_1.name);

// @api: error.type_error
// @expect: type fail
// @expect: TypeError
const e_error_type_error_2 = new TypeError("type fail");
console.log(e_error_type_error_2.message);
console.log(e_error_type_error_2.name);

// @api: error.syntax_error
// @expect: syntax fail
// @expect: SyntaxError
const e_error_syntax_error_3 = new SyntaxError("syntax fail");
console.log(e_error_syntax_error_3.message);
console.log(e_error_syntax_error_3.name);
