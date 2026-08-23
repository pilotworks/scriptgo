// ScriptGo Corpus: Process Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: process.argv
// @expect: true
console.log(process.argv.length > 0);

// @api: process.cwd
// @expect: string
console.log(typeof process.cwd());

// @api: process.env
// @expect: true
console.log(typeof process.env === "object");

// @api: process.exit
// @expect: exit defined
console.log("exit defined");

