// ScriptGo Corpus: Performance Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: Performance.now
// @expect: true
const t1: number = performance.now();
const t2: number = performance.now();
console.log(t2 >= t1);
