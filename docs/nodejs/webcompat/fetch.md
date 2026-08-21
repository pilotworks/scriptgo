# Fetch API & Headers / Request / Response Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG Fetch API & Headers / Request / Response Specification](https://wintercg.org/)  
> **Type Definition Source**: [microsoft/TypeScript lib.dom.d.ts (Server subset)](https://github.com/microsoft/TypeScript/blob/main/src/lib/lib.dom.d.ts)  
> **Gate Oracle**: Web Platform Tests (WPT) & Node.js 22 LTS WPT runner

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
| `Headers` | `(...) => any` | `__fetch.Headers` | ✅ Done | `internal/compiler/testdata/corpus/fetch/headers/` |
| `AbortController` | `(...) => any` | `__fetch.AbortController` | 📋 Planned | - |
| `AbortSignal` | `(...) => any` | `__fetch.AbortSignal` | 📋 Planned | - |
| `Blob` | `(...) => any` | `__fetch.Blob` | 📋 Planned | - |
| `BroadcastChannel` | `(...) => any` | `__fetch.BroadcastChannel` | 📋 Planned | - |
| `Buffer` | `(...) => any` | `__fetch.Buffer` | 📋 Planned | - |
| `ByteLengthQueuingStrategy` | `(...) => any` | `__fetch.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__fetch.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__fetch.CountQueuingStrategy` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__fetch.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__fetch.CryptoKey` | 📋 Planned | - |
| `CustomEvent` | `(...) => any` | `__fetch.CustomEvent` | 📋 Planned | - |
| `DOMException` | `(...) => any` | `__fetch.DOMException` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__fetch.DecompressionStream` | 📋 Planned | - |
| `Event` | `(...) => any` | `__fetch.Event` | 📋 Planned | - |
| `EventSource` | `(...) => any` | `__fetch.EventSource` | 📋 Planned | - |
| `EventTarget` | `(...) => any` | `__fetch.EventTarget` | 📋 Planned | - |
| `File` | `(...) => any` | `__fetch.File` | 📋 Planned | - |
| `FormData` | `(...) => any` | `__fetch.FormData` | 📋 Planned | - |
| `MessageChannel` | `(...) => any` | `__fetch.MessageChannel` | 📋 Planned | - |
| `MessageEvent` | `(...) => any` | `__fetch.MessageEvent` | 📋 Planned | - |
| `MessagePort` | `(...) => any` | `__fetch.MessagePort` | 📋 Planned | - |
| `Navigator` | `(...) => any` | `__fetch.Navigator` | 📋 Planned | - |
| `PerformanceEntry` | `(...) => any` | `__fetch.PerformanceEntry` | 📋 Planned | - |
| `PerformanceMark` | `(...) => any` | `__fetch.PerformanceMark` | 📋 Planned | - |
| `PerformanceMeasure` | `(...) => any` | `__fetch.PerformanceMeasure` | 📋 Planned | - |
| `PerformanceObserver` | `(...) => any` | `__fetch.PerformanceObserver` | 📋 Planned | - |
| `PerformanceObserverEntryList` | `(...) => any` | `__fetch.PerformanceObserverEntryList` | 📋 Planned | - |
| `PerformanceResourceTiming` | `(...) => any` | `__fetch.PerformanceResourceTiming` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__fetch.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__fetch.ReadableStream` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__fetch.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__fetch.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__fetch.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__fetch.ReadableStreamDefaultReader` | 📋 Planned | - |
| `Request` | `(...) => any` | `__fetch.Request` | 📋 Planned | - |
| `Response` | `(...) => any` | `__fetch.Response` | 📋 Planned | - |
| `Storage` | `(...) => any` | `__fetch.Storage` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__fetch.SubtleCrypto` | 📋 Planned | - |
| `TextDecoder` | `(...) => any` | `__fetch.TextDecoder` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__fetch.TextDecoderStream` | 📋 Planned | - |
| `TextEncoder` | `(...) => any` | `__fetch.TextEncoder` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__fetch.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__fetch.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__fetch.TransformStreamDefaultController` | 📋 Planned | - |
| `URL` | `(...) => any` | `__fetch.URL` | 📋 Planned | - |
| `URLSearchParams` | `(...) => any` | `__fetch.URLSearchParams` | 📋 Planned | - |
| `WebAssembly` | `(...) => any` | `__fetch.WebAssembly` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__fetch.WebSocket` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__fetch.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__fetch.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__fetch.WritableStreamDefaultWriter` | 📋 Planned | - |
| `abortController.abort([reason])` | `(...) => any` | `__fetch.abortController.abort` | 📋 Planned | - |
| `abortSignal.throwIfAborted()` | `(...) => any` | `__fetch.abortSignal.throwIfAborted` | 📋 Planned | - |
| `aborted` | `any` | `__fetch.aborted` | 📋 Planned | - |
| `atob(data)` | `(...) => any` | `__fetch.atob` | 📋 Planned | - |
| `btoa(data)` | `(...) => any` | `__fetch.btoa` | 📋 Planned | - |
| `clearImmediate(immediateObject)` | `(...) => any` | `__fetch.clearImmediate` | 📋 Planned | - |
| `clearInterval(intervalObject)` | `(...) => any` | `__fetch.clearInterval` | 📋 Planned | - |
| `clearTimeout(timeoutObject)` | `(...) => any` | `__fetch.clearTimeout` | 📋 Planned | - |
| `onabort` | `any` | `__fetch.onabort` | 📋 Planned | - |
| `queueMicrotask(callback)` | `(...) => any` | `__fetch.queueMicrotask` | 📋 Planned | - |
| `reason` | `any` | `__fetch.reason` | 📋 Planned | - |
| `require()` | `(...) => any` | `__fetch.require` | 📋 Planned | - |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__fetch.setImmediate` | 📋 Planned | - |
| `setInterval(callback, delay[, ...args])` | `(...) => any` | `__fetch.setInterval` | 📋 Planned | - |
| `setTimeout(callback, delay[, ...args])` | `(...) => any` | `__fetch.setTimeout` | 📋 Planned | - |
| `signal` | `any` | `__fetch.signal` | 📋 Planned | - |
| `structuredClone(value[, options])` | `(...) => any` | `__fetch.structuredClone` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `fetch` are organized per API under `internal/compiler/testdata/corpus/fetch/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/fetch/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
