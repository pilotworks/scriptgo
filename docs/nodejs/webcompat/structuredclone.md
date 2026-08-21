# structuredClone API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG structuredClone API Specification](https://wintercg.org/)  
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
| `AbortController` | `(...) => any` | `__structuredclone.AbortController` | 📋 Planned | - |
| `AbortSignal` | `(...) => any` | `__structuredclone.AbortSignal` | 📋 Planned | - |
| `Blob` | `(...) => any` | `__structuredclone.Blob` | 📋 Planned | - |
| `BroadcastChannel` | `(...) => any` | `__structuredclone.BroadcastChannel` | 📋 Planned | - |
| `Buffer` | `(...) => any` | `__structuredclone.Buffer` | 📋 Planned | - |
| `ByteLengthQueuingStrategy` | `(...) => any` | `__structuredclone.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__structuredclone.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__structuredclone.CountQueuingStrategy` | 📋 Planned | - |
| `Crypto` | `(...) => any` | `__structuredclone.Crypto` | 📋 Planned | - |
| `CryptoKey` | `(...) => any` | `__structuredclone.CryptoKey` | 📋 Planned | - |
| `CustomEvent` | `(...) => any` | `__structuredclone.CustomEvent` | 📋 Planned | - |
| `DOMException` | `(...) => any` | `__structuredclone.DOMException` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__structuredclone.DecompressionStream` | 📋 Planned | - |
| `Event` | `(...) => any` | `__structuredclone.Event` | 📋 Planned | - |
| `EventSource` | `(...) => any` | `__structuredclone.EventSource` | 📋 Planned | - |
| `EventTarget` | `(...) => any` | `__structuredclone.EventTarget` | 📋 Planned | - |
| `File` | `(...) => any` | `__structuredclone.File` | 📋 Planned | - |
| `FormData` | `(...) => any` | `__structuredclone.FormData` | 📋 Planned | - |
| `Headers` | `(...) => any` | `__structuredclone.Headers` | 📋 Planned | - |
| `MessageChannel` | `(...) => any` | `__structuredclone.MessageChannel` | 📋 Planned | - |
| `MessageEvent` | `(...) => any` | `__structuredclone.MessageEvent` | 📋 Planned | - |
| `MessagePort` | `(...) => any` | `__structuredclone.MessagePort` | 📋 Planned | - |
| `Navigator` | `(...) => any` | `__structuredclone.Navigator` | 📋 Planned | - |
| `PerformanceEntry` | `(...) => any` | `__structuredclone.PerformanceEntry` | 📋 Planned | - |
| `PerformanceMark` | `(...) => any` | `__structuredclone.PerformanceMark` | 📋 Planned | - |
| `PerformanceMeasure` | `(...) => any` | `__structuredclone.PerformanceMeasure` | 📋 Planned | - |
| `PerformanceObserver` | `(...) => any` | `__structuredclone.PerformanceObserver` | 📋 Planned | - |
| `PerformanceObserverEntryList` | `(...) => any` | `__structuredclone.PerformanceObserverEntryList` | 📋 Planned | - |
| `PerformanceResourceTiming` | `(...) => any` | `__structuredclone.PerformanceResourceTiming` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__structuredclone.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__structuredclone.ReadableStream` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__structuredclone.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__structuredclone.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__structuredclone.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__structuredclone.ReadableStreamDefaultReader` | 📋 Planned | - |
| `Request` | `(...) => any` | `__structuredclone.Request` | 📋 Planned | - |
| `Response` | `(...) => any` | `__structuredclone.Response` | 📋 Planned | - |
| `Storage` | `(...) => any` | `__structuredclone.Storage` | 📋 Planned | - |
| `SubtleCrypto` | `(...) => any` | `__structuredclone.SubtleCrypto` | 📋 Planned | - |
| `TextDecoder` | `(...) => any` | `__structuredclone.TextDecoder` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__structuredclone.TextDecoderStream` | 📋 Planned | - |
| `TextEncoder` | `(...) => any` | `__structuredclone.TextEncoder` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__structuredclone.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__structuredclone.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__structuredclone.TransformStreamDefaultController` | 📋 Planned | - |
| `URL` | `(...) => any` | `__structuredclone.URL` | 📋 Planned | - |
| `URLSearchParams` | `(...) => any` | `__structuredclone.URLSearchParams` | 📋 Planned | - |
| `WebAssembly` | `(...) => any` | `__structuredclone.WebAssembly` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__structuredclone.WebSocket` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__structuredclone.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__structuredclone.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__structuredclone.WritableStreamDefaultWriter` | 📋 Planned | - |
| `abortController.abort([reason])` | `(...) => any` | `__structuredclone.abortController.abort` | 📋 Planned | - |
| `abortSignal.throwIfAborted()` | `(...) => any` | `__structuredclone.abortSignal.throwIfAborted` | 📋 Planned | - |
| `aborted` | `any` | `__structuredclone.aborted` | 📋 Planned | - |
| `atob(data)` | `(...) => any` | `__structuredclone.atob` | 📋 Planned | - |
| `btoa(data)` | `(...) => any` | `__structuredclone.btoa` | 📋 Planned | - |
| `clearImmediate(immediateObject)` | `(...) => any` | `__structuredclone.clearImmediate` | 📋 Planned | - |
| `clearInterval(intervalObject)` | `(...) => any` | `__structuredclone.clearInterval` | 📋 Planned | - |
| `clearTimeout(timeoutObject)` | `(...) => any` | `__structuredclone.clearTimeout` | 📋 Planned | - |
| `onabort` | `any` | `__structuredclone.onabort` | 📋 Planned | - |
| `queueMicrotask(callback)` | `(...) => any` | `__structuredclone.queueMicrotask` | 📋 Planned | - |
| `reason` | `any` | `__structuredclone.reason` | 📋 Planned | - |
| `require()` | `(...) => any` | `__structuredclone.require` | 📋 Planned | - |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__structuredclone.setImmediate` | 📋 Planned | - |
| `setInterval(callback, delay[, ...args])` | `(...) => any` | `__structuredclone.setInterval` | 📋 Planned | - |
| `setTimeout(callback, delay[, ...args])` | `(...) => any` | `__structuredclone.setTimeout` | 📋 Planned | - |
| `signal` | `any` | `__structuredclone.signal` | 📋 Planned | - |
| `structuredClone(value[, options])` | `(...) => any` | `__structuredclone.structuredClone` | 📋 Planned | - |

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
