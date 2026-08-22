// ScriptGo Corpus: Perf_hooks Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: perf_hooks.now
// @expect: true
const t1_perf_hooks_now_0: number = performance.now();
const t2_perf_hooks_now_0: number = performance.now();
console.log(t2_perf_hooks_now_0 >= t1_perf_hooks_now_0);

// @api: perf_hooks.performance
// @expect: true
console.log(performance.now() >= 0);
