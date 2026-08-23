// ScriptGo Corpus: SuppressedError Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: SuppressedError
// @api: SuppressedError.constructor
// @api: SuppressedError.error
// @api: SuppressedError.suppressed
// @api: SuppressedError.message
// @api: SuppressedError.name
// @expect: primary error
// @expect: suppressed error
// @expect: an error was suppressed
// @expect: SuppressedError
const se = new SuppressedError("primary error", "suppressed error", "an error was suppressed");
console.log(se.error);
console.log(se.suppressed);
console.log(se.message);
console.log(se.name);
