// ScriptGo Corpus: Process Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as process from "node:process";

// @api: process.argv
// @expect: true
console.log(process.argv.length > 0);

// @api: process.cwd
// @expect: string
console.log(typeof process.cwd());
