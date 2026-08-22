// ScriptGo Corpus: Async Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: async.queueMicrotask
// @expect: synchronous log
// @expect: microtask ran
queueMicrotask(() => {
    console.log("microtask ran");
});
console.log("synchronous log");
