# Global Objects & Identifiers Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Global Objects & Identifiers Documentation](https://nodejs.org/docs/latest-v22.x/api/globals.html)  
> **Type Definition Source**: [@types/node/globals.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-global-*.js, test-globals-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `AbortController` | `(...) => any` | `__globals.AbortController` | 📋 Planned | - |
| `AbortSignal` | `(...) => any` | `__globals.AbortSignal` | 📋 Planned | - |
| `Blob` | `(...) => any` | `__globals.Blob` | 📋 Planned | - |
| `BroadcastChannel` | `(...) => any` | `__globals.BroadcastChannel` | 📋 Planned | - |
| `Buffer` | `(...) => any` | `__globals.Buffer` | 📋 Planned | - |
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
| `File` | `(...) => any` | `__globals.File` | 📋 Planned | - |
| `FormData` | `(...) => any` | `__globals.FormData` | 📋 Planned | - |
| `Headers` | `(...) => any` | `__globals.Headers` | 📋 Planned | - |
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
| `Response` | `(...) => any` | `__globals.Response` | 📋 Planned | - |
| `Storage` | `(...) => any` | `__globals.Storage` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__globals.SubtleCrypto` | 📋 Planned | - |
| `TextDecoder` | `(...) => any` | `__globals.TextDecoder` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__globals.TextDecoderStream` | 📋 Planned | - |
| `TextEncoder` | `(...) => any` | `__globals.TextEncoder` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__globals.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__globals.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__globals.TransformStreamDefaultController` | 📋 Planned | - |
| `URL` | `(...) => any` | `__globals.URL` | 📋 Planned | - |
| `URLSearchParams` | `(...) => any` | `__globals.URLSearchParams` | 📋 Planned | - |
| `WebAssembly` | `(...) => any` | `__globals.WebAssembly` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__globals.WebSocket` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__globals.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__globals.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__globals.WritableStreamDefaultWriter` | 📋 Planned | - |
| `abortController.abort([reason])` | `(...) => any` | `__globals.abortController.abort` | 📋 Planned | - |
| `abortSignal.throwIfAborted()` | `(...) => any` | `__globals.abortSignal.throwIfAborted` | 📋 Planned | - |
| `aborted` | `any` | `__globals.aborted` | 📋 Planned | - |
| `atob(data)` | `(...) => any` | `__globals.atob` | 📋 Planned | - |
| `btoa(data)` | `(...) => any` | `__globals.btoa` | 📋 Planned | - |
| `clearImmediate(immediateObject)` | `(...) => any` | `__globals.clearImmediate` | 📋 Planned | - |
| `clearInterval(intervalObject)` | `(...) => any` | `__globals.clearInterval` | 📋 Planned | - |
| `clearTimeout(timeoutObject)` | `(...) => any` | `__globals.clearTimeout` | 📋 Planned | - |
| `onabort` | `any` | `__globals.onabort` | 📋 Planned | - |
| `queueMicrotask(callback)` | `(...) => any` | `__globals.queueMicrotask` | 📋 Planned | - |
| `reason` | `any` | `__globals.reason` | 📋 Planned | - |
| `require()` | `(...) => any` | `__globals.require` | 📋 Planned | - |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__globals.setImmediate` | 📋 Planned | - |
| `setInterval(callback, delay[, ...args])` | `(...) => any` | `__globals.setInterval` | 📋 Planned | - |
| `setTimeout(callback, delay[, ...args])` | `(...) => any` | `__globals.setTimeout` | 📋 Planned | - |
| `signal` | `any` | `__globals.signal` | 📋 Planned | - |
| `structuredClone(value[, options])` | `(...) => any` | `__globals.structuredClone` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `language` are organized per API under `internal/compiler/testdata/corpus/language/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/language/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
