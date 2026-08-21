# Performance Hooks (node:perf_hooks) Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:perf_hooks`  
> **Specification Reference**: [Node.js 22 LTS Performance Hooks (node:perf_hooks) API Documentation](https://nodejs.org/docs/latest-v22.x/api/perf_hooks.html)  
> **Type Definition Source**: [@types/node/perf_hooks.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-perf_hooks-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:perf_hooks`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `perf_hooks.performance` | `any` | `__perf_hooks.perf_hooks.performance` | ✅ Done | `internal/compiler/testdata/corpus/perf_hooks/performance/` |
| `Histogram` | `(...) => any` | `__perf_hooks.Histogram` | 📋 Planned | - |
| `PerformanceEntry` | `(...) => any` | `__perf_hooks.PerformanceEntry` | 📋 Planned | - |
| `PerformanceMark` | `(...) => any` | `__perf_hooks.PerformanceMark` | 📋 Planned | - |
| `PerformanceMeasure` | `(...) => any` | `__perf_hooks.PerformanceMeasure` | 📋 Planned | - |
| `PerformanceNodeEntry` | `(...) => any` | `__perf_hooks.PerformanceNodeEntry` | 📋 Planned | - |
| `PerformanceNodeTiming` | `(...) => any` | `__perf_hooks.PerformanceNodeTiming` | 📋 Planned | - |
| `PerformanceObserver` | `(...) => any` | `__perf_hooks.PerformanceObserver` | 📋 Planned | - |
| `PerformanceObserverEntryList` | `(...) => any` | `__perf_hooks.PerformanceObserverEntryList` | 📋 Planned | - |
| `PerformanceResourceTiming` | `(...) => any` | `__perf_hooks.PerformanceResourceTiming` | 📋 Planned | - |
| `bootstrapComplete` | `any` | `__perf_hooks.bootstrapComplete` | 📋 Planned | - |
| `connectEnd` | `any` | `__perf_hooks.connectEnd` | 📋 Planned | - |
| `connectStart` | `any` | `__perf_hooks.connectStart` | 📋 Planned | - |
| `count` | `any` | `__perf_hooks.count` | 📋 Planned | - |
| `countBigInt` | `any` | `__perf_hooks.countBigInt` | 📋 Planned | - |
| `decodedBodySize` | `any` | `__perf_hooks.decodedBodySize` | 📋 Planned | - |
| `detail` | `any` | `__perf_hooks.detail` | 📋 Planned | - |
| `domainLookupEnd` | `any` | `__perf_hooks.domainLookupEnd` | 📋 Planned | - |
| `domainLookupStart` | `any` | `__perf_hooks.domainLookupStart` | 📋 Planned | - |
| `duration` | `any` | `__perf_hooks.duration` | 📋 Planned | - |
| `encodedBodySize` | `any` | `__perf_hooks.encodedBodySize` | 📋 Planned | - |
| `entryType` | `any` | `__perf_hooks.entryType` | 📋 Planned | - |
| `environment` | `any` | `__perf_hooks.environment` | 📋 Planned | - |
| `exceeds` | `any` | `__perf_hooks.exceeds` | 📋 Planned | - |
| `exceedsBigInt` | `any` | `__perf_hooks.exceedsBigInt` | 📋 Planned | - |
| `fetchStart` | `any` | `__perf_hooks.fetchStart` | 📋 Planned | - |
| `flags` | `any` | `__perf_hooks.flags` | 📋 Planned | - |
| `histogram.add(other)` | `(...) => any` | `__perf_hooks.histogram.add` | 📋 Planned | - |
| `histogram.disable()` | `(...) => any` | `__perf_hooks.histogram.disable` | 📋 Planned | - |
| `histogram.enable()` | `(...) => any` | `__perf_hooks.histogram.enable` | 📋 Planned | - |
| `histogram.percentile(percentile)` | `(...) => any` | `__perf_hooks.histogram.percentile` | 📋 Planned | - |
| `histogram.percentileBigInt(percentile)` | `(...) => any` | `__perf_hooks.histogram.percentileBigInt` | 📋 Planned | - |
| `histogram.record(val)` | `(...) => any` | `__perf_hooks.histogram.record` | 📋 Planned | - |
| `histogram.recordDelta()` | `(...) => any` | `__perf_hooks.histogram.recordDelta` | 📋 Planned | - |
| `histogram.reset()` | `(...) => any` | `__perf_hooks.histogram.reset` | 📋 Planned | - |
| `idleTime` | `any` | `__perf_hooks.idleTime` | 📋 Planned | - |
| `kind` | `any` | `__perf_hooks.kind` | 📋 Planned | - |
| `loopExit` | `any` | `__perf_hooks.loopExit` | 📋 Planned | - |
| `loopStart` | `any` | `__perf_hooks.loopStart` | 📋 Planned | - |
| `max` | `any` | `__perf_hooks.max` | 📋 Planned | - |
| `maxBigInt` | `any` | `__perf_hooks.maxBigInt` | 📋 Planned | - |
| `mean` | `any` | `__perf_hooks.mean` | 📋 Planned | - |
| `min` | `any` | `__perf_hooks.min` | 📋 Planned | - |
| `minBigInt` | `any` | `__perf_hooks.minBigInt` | 📋 Planned | - |
| `name` | `any` | `__perf_hooks.name` | 📋 Planned | - |
| `nodeStart` | `any` | `__perf_hooks.nodeStart` | 📋 Planned | - |
| `percentiles` | `any` | `__perf_hooks.percentiles` | 📋 Planned | - |
| `percentilesBigInt` | `any` | `__perf_hooks.percentilesBigInt` | 📋 Planned | - |
| `perf_hooks.createHistogram([options])` | `(...) => any` | `__perf_hooks.perf_hooks.createHistogram` | 📋 Planned | - |
| `perf_hooks.monitorEventLoopDelay([options])` | `(...) => any` | `__perf_hooks.perf_hooks.monitorEventLoopDelay` | 📋 Planned | - |
| `performanceObserver.disconnect()` | `(...) => any` | `__perf_hooks.performanceObserver.disconnect` | 📋 Planned | - |
| `performanceObserver.observe(options)` | `(...) => any` | `__perf_hooks.performanceObserver.observe` | 📋 Planned | - |
| `performanceObserver.takeRecords()` | `(...) => any` | `__perf_hooks.performanceObserver.takeRecords` | 📋 Planned | - |
| `performanceObserverEntryList.getEntries()` | `(...) => any` | `__perf_hooks.performanceObserverEntryList.getEntries` | 📋 Planned | - |
| `performanceObserverEntryList.getEntriesByName(name[, type])` | `(...) => any` | `__perf_hooks.performanceObserverEntryList.getEntriesByName` | 📋 Planned | - |
| `performanceObserverEntryList.getEntriesByType(type)` | `(...) => any` | `__perf_hooks.performanceObserverEntryList.getEntriesByType` | 📋 Planned | - |
| `performanceResourceTiming.toJSON()` | `(...) => any` | `__perf_hooks.performanceResourceTiming.toJSON` | 📋 Planned | - |
| `redirectEnd` | `any` | `__perf_hooks.redirectEnd` | 📋 Planned | - |
| `redirectStart` | `any` | `__perf_hooks.redirectStart` | 📋 Planned | - |
| `requestStart` | `any` | `__perf_hooks.requestStart` | 📋 Planned | - |
| `responseEnd` | `any` | `__perf_hooks.responseEnd` | 📋 Planned | - |
| `secureConnectionStart` | `any` | `__perf_hooks.secureConnectionStart` | 📋 Planned | - |
| `startTime` | `any` | `__perf_hooks.startTime` | 📋 Planned | - |
| `stddev` | `any` | `__perf_hooks.stddev` | 📋 Planned | - |
| `supportedEntryTypes` | `any` | `__perf_hooks.supportedEntryTypes` | 📋 Planned | - |
| `transferSize` | `any` | `__perf_hooks.transferSize` | 📋 Planned | - |
| `uvMetricsInfo` Returns: {Object}` | `any` | `__perf_hooks.uvMetricsInfo` Returns: {Object}` | 📋 Planned | - |
| `v8Start` | `any` | `__perf_hooks.v8Start` | 📋 Planned | - |
| `workerStart` | `any` | `__perf_hooks.workerStart` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `perf_hooks` are organized per API under `internal/compiler/testdata/corpus/perf_hooks/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/perf_hooks/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
