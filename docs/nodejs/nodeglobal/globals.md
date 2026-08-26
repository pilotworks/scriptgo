# Global Objects Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Globals Documentation](https://nodejs.org/docs/latest-v22.x/api/globals.html)  
> **Type Definition Source**: [@types/node/globals.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-global-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `Blob` | `(...) => any` | `__globals.Blob` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer` | `(...) => any` | `__globals.Buffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `File` | `(...) => any` | `__globals.File` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Headers` | `(...) => any` | `__globals.Headers` | ✅ Done | `internal/compiler/testdata/corpus/api/fetch.ts` |
| `Response` | `(...) => any` | `__globals.Response` | ✅ Done | `internal/compiler/testdata/corpus/api/fetch.ts` |
| `URL` | `(...) => any` | `__globals.URL` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URLSearchParams` | `(...) => any` | `__globals.URLSearchParams` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `abortController.abort([reason])` | `(...) => any` | `__globals.abortController.abort` | ✅ Done | `internal/compiler/testdata/corpus/api/abortcontroller.ts` |
| `abortSignal.throwIfAborted()` | `(...) => any` | `__globals.abortSignal.throwIfAborted` | ✅ Done | `internal/compiler/testdata/corpus/api/abortcontroller.ts` |
| `atob(data)` | `(...) => any` | `__globals.atob` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `btoa(data)` | `(...) => any` | `__globals.btoa` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `clearImmediate(immediateObject)` | `(...) => any` | `__globals.clearImmediate` | ✅ Done | `internal/compiler/testdata/corpus/api/timers.ts` |
| `clearTimeout(timeoutObject)` | `(...) => any` | `__globals.clearTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/timers.ts` |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__globals.setImmediate` | ✅ Done | `internal/compiler/testdata/corpus/api/timers.ts` |
| `setInterval(callback, delay[, ...args])` | `(...) => any` | `__globals.setInterval` | ✅ Done | `internal/compiler/testdata/corpus/api/timers.ts` |
| `setTimeout(callback, delay[, ...args])` | `(...) => any` | `__globals.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/timers.ts` |
| `AbortController` | `(...) => any` | `__globals.AbortController` | 📋 Planned | - |
| `AbortSignal` | `(...) => any` | `__globals.AbortSignal` | 📋 Planned | - |
| `BroadcastChannel` | `(...) => any` | `__globals.BroadcastChannel` | 📋 Planned | - |
| `ByteLengthQueuingStrategy` | `(...) => any` | `__globals.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__globals.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__globals.CountQueuingStrategy` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__globals.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__globals.CryptoKey` | 📋 Planned | - |
| `CustomEvent` | `(...) => any` | `__globals.CustomEvent` | 📋 Planned | - |
| `DOMException` | `(...) => any` | `__globals.DOMException` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__globals.DecompressionStream` | 📋 Planned | - |
| `Event` | `(...) => any` | `__globals.Event` | 📋 Planned | - |
| `EventSource` | `(...) => any` | `__globals.EventSource` | 📋 Planned | - |
| `EventTarget` | `(...) => any` | `__globals.EventTarget` | 📋 Planned | - |
| `FormData` | `(...) => any` | `__globals.FormData` | 📋 Planned | - |
| `MessageChannel` | `(...) => any` | `__globals.MessageChannel` | 📋 Planned | - |
| `MessageEvent` | `(...) => any` | `__globals.MessageEvent` | 📋 Planned | - |
| `MessagePort` | `(...) => any` | `__globals.MessagePort` | 📋 Planned | - |
| `Navigator` | `(...) => any` | `__globals.Navigator` | 📋 Planned | - |
| `PerformanceEntry` | `(...) => any` | `__globals.PerformanceEntry` | 📋 Planned | - |
| `PerformanceMark` | `(...) => any` | `__globals.PerformanceMark` | 📋 Planned | - |
| `PerformanceMeasure` | `(...) => any` | `__globals.PerformanceMeasure` | 📋 Planned | - |
| `PerformanceObserver` | `(...) => any` | `__globals.PerformanceObserver` | 📋 Planned | - |
| `PerformanceObserverEntryList` | `(...) => any` | `__globals.PerformanceObserverEntryList` | 📋 Planned | - |
| `PerformanceResourceTiming` | `(...) => any` | `__globals.PerformanceResourceTiming` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__globals.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__globals.ReadableStream` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__globals.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__globals.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__globals.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__globals.ReadableStreamDefaultReader` | 📋 Planned | - |
| `Request` | `(...) => any` | `__globals.Request` | 📋 Planned | - |
| `Storage` | `(...) => any` | `__globals.Storage` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__globals.SubtleCrypto` | 📋 Planned | - |
| `TextDecoder` | `(...) => any` | `__globals.TextDecoder` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__globals.TextDecoderStream` | 📋 Planned | - |
| `TextEncoder` | `(...) => any` | `__globals.TextEncoder` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__globals.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__globals.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__globals.TransformStreamDefaultController` | 📋 Planned | - |
| `WebAssembly` | `(...) => any` | `__globals.WebAssembly` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__globals.WebSocket` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__globals.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__globals.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__globals.WritableStreamDefaultWriter` | 📋 Planned | - |
| `aborted` | `any` | `__globals.aborted` | 📋 Planned | - |
| `clearInterval(intervalObject)` | `(...) => any` | `__globals.clearInterval` | 📋 Planned | - |
| `onabort` | `any` | `__globals.onabort` | 📋 Planned | - |
| `queueMicrotask(callback)` | `(...) => any` | `__globals.queueMicrotask` | 📋 Planned | - |
| `reason` | `any` | `__globals.reason` | 📋 Planned | - |
| `require()` | `(...) => any` | `__globals.require` | 📋 Planned | - |
| `signal` | `any` | `__globals.signal` | 📋 Planned | - |
| `structuredClone(value[, options])` | `(...) => any` | `__globals.structuredClone` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `globals` are organized per API under `internal/compiler/testdata/corpus/globals/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/globals/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
