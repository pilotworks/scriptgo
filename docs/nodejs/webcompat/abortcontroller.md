# AbortController & AbortSignal Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG AbortController & AbortSignal Specification](https://wintercg.org/)  
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
| `AbortController` | `(...) => any` | `__abortcontroller.AbortController` | 📋 Planned | - |
| `AbortSignal` | `(...) => any` | `__abortcontroller.AbortSignal` | 📋 Planned | - |
| `Blob` | `(...) => any` | `__abortcontroller.Blob` | 📋 Planned | - |
| `BroadcastChannel` | `(...) => any` | `__abortcontroller.BroadcastChannel` | 📋 Planned | - |
| `Buffer` | `(...) => any` | `__abortcontroller.Buffer` | 📋 Planned | - |
| `ByteLengthQueuingStrategy` | `(...) => any` | `__abortcontroller.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__abortcontroller.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__abortcontroller.CountQueuingStrategy` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__abortcontroller.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__abortcontroller.CryptoKey` | 📋 Planned | - |
| `CustomEvent` | `(...) => any` | `__abortcontroller.CustomEvent` | 📋 Planned | - |
| `DOMException` | `(...) => any` | `__abortcontroller.DOMException` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__abortcontroller.DecompressionStream` | 📋 Planned | - |
| `Event` | `(...) => any` | `__abortcontroller.Event` | 📋 Planned | - |
| `EventSource` | `(...) => any` | `__abortcontroller.EventSource` | 📋 Planned | - |
| `EventTarget` | `(...) => any` | `__abortcontroller.EventTarget` | 📋 Planned | - |
| `File` | `(...) => any` | `__abortcontroller.File` | 📋 Planned | - |
| `FormData` | `(...) => any` | `__abortcontroller.FormData` | 📋 Planned | - |
| `Headers` | `(...) => any` | `__abortcontroller.Headers` | 📋 Planned | - |
| `MessageChannel` | `(...) => any` | `__abortcontroller.MessageChannel` | 📋 Planned | - |
| `MessageEvent` | `(...) => any` | `__abortcontroller.MessageEvent` | 📋 Planned | - |
| `MessagePort` | `(...) => any` | `__abortcontroller.MessagePort` | 📋 Planned | - |
| `Navigator` | `(...) => any` | `__abortcontroller.Navigator` | 📋 Planned | - |
| `PerformanceEntry` | `(...) => any` | `__abortcontroller.PerformanceEntry` | 📋 Planned | - |
| `PerformanceMark` | `(...) => any` | `__abortcontroller.PerformanceMark` | 📋 Planned | - |
| `PerformanceMeasure` | `(...) => any` | `__abortcontroller.PerformanceMeasure` | 📋 Planned | - |
| `PerformanceObserver` | `(...) => any` | `__abortcontroller.PerformanceObserver` | 📋 Planned | - |
| `PerformanceObserverEntryList` | `(...) => any` | `__abortcontroller.PerformanceObserverEntryList` | 📋 Planned | - |
| `PerformanceResourceTiming` | `(...) => any` | `__abortcontroller.PerformanceResourceTiming` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__abortcontroller.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__abortcontroller.ReadableStream` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__abortcontroller.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__abortcontroller.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__abortcontroller.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__abortcontroller.ReadableStreamDefaultReader` | 📋 Planned | - |
| `Request` | `(...) => any` | `__abortcontroller.Request` | 📋 Planned | - |
| `Response` | `(...) => any` | `__abortcontroller.Response` | 📋 Planned | - |
| `Storage` | `(...) => any` | `__abortcontroller.Storage` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__abortcontroller.SubtleCrypto` | 📋 Planned | - |
| `TextDecoder` | `(...) => any` | `__abortcontroller.TextDecoder` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__abortcontroller.TextDecoderStream` | 📋 Planned | - |
| `TextEncoder` | `(...) => any` | `__abortcontroller.TextEncoder` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__abortcontroller.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__abortcontroller.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__abortcontroller.TransformStreamDefaultController` | 📋 Planned | - |
| `URL` | `(...) => any` | `__abortcontroller.URL` | 📋 Planned | - |
| `URLSearchParams` | `(...) => any` | `__abortcontroller.URLSearchParams` | 📋 Planned | - |
| `WebAssembly` | `(...) => any` | `__abortcontroller.WebAssembly` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__abortcontroller.WebSocket` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__abortcontroller.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__abortcontroller.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__abortcontroller.WritableStreamDefaultWriter` | 📋 Planned | - |
| `abortController.abort([reason])` | `(...) => any` | `__abortcontroller.abortController.abort` | 📋 Planned | - |
| `abortSignal.throwIfAborted()` | `(...) => any` | `__abortcontroller.abortSignal.throwIfAborted` | 📋 Planned | - |
| `aborted` | `any` | `__abortcontroller.aborted` | 📋 Planned | - |
| `atob(data)` | `(...) => any` | `__abortcontroller.atob` | 📋 Planned | - |
| `btoa(data)` | `(...) => any` | `__abortcontroller.btoa` | 📋 Planned | - |
| `clearImmediate(immediateObject)` | `(...) => any` | `__abortcontroller.clearImmediate` | 📋 Planned | - |
| `clearInterval(intervalObject)` | `(...) => any` | `__abortcontroller.clearInterval` | 📋 Planned | - |
| `clearTimeout(timeoutObject)` | `(...) => any` | `__abortcontroller.clearTimeout` | 📋 Planned | - |
| `onabort` | `any` | `__abortcontroller.onabort` | 📋 Planned | - |
| `queueMicrotask(callback)` | `(...) => any` | `__abortcontroller.queueMicrotask` | 📋 Planned | - |
| `reason` | `any` | `__abortcontroller.reason` | 📋 Planned | - |
| `require()` | `(...) => any` | `__abortcontroller.require` | 📋 Planned | - |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__abortcontroller.setImmediate` | 📋 Planned | - |
| `setInterval(callback, delay[, ...args])` | `(...) => any` | `__abortcontroller.setInterval` | 📋 Planned | - |
| `setTimeout(callback, delay[, ...args])` | `(...) => any` | `__abortcontroller.setTimeout` | 📋 Planned | - |
| `signal` | `any` | `__abortcontroller.signal` | 📋 Planned | - |
| `structuredClone(value[, options])` | `(...) => any` | `__abortcontroller.structuredClone` | 📋 Planned | - |

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
