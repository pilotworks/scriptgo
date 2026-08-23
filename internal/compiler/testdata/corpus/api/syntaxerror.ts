// ScriptGo Corpus: SyntaxError Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: SyntaxError
// @api: SyntaxError.constructor
// @api: SyntaxError.message
// @api: SyntaxError.name
// @expect: unexpected token
// @expect: SyntaxError
const synErr = new SyntaxError("unexpected token");
console.log(synErr.message);
console.log(synErr.name);
