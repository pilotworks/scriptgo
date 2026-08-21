# WHATWG Web Streams API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG WHATWG Web Streams API Specification](https://wintercg.org/)  
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
| `ByteLengthQueuingStrategy` | `(...) => any` | `__streams.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__streams.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__streams.CountQueuingStrategy` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__streams.DecompressionStream` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__streams.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__streams.ReadableStream` | 📋 Planned | - |
| `ReadableStream.from(iterable)` | `(...) => any` | `__streams.ReadableStream.from` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__streams.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__streams.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__streams.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__streams.ReadableStreamDefaultReader` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__streams.TextDecoderStream` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__streams.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__streams.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__streams.TransformStreamDefaultController` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__streams.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__streams.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__streams.WritableStreamDefaultWriter` | 📋 Planned | - |
| `byobRequest` | `any` | `__streams.byobRequest` | 📋 Planned | - |
| `closed` | `any` | `__streams.closed` | 📋 Planned | - |
| `desiredSize` | `any` | `__streams.desiredSize` | 📋 Planned | - |
| `encoding` | `any` | `__streams.encoding` | 📋 Planned | - |
| `fatal` | `any` | `__streams.fatal` | 📋 Planned | - |
| `highWaterMark` | `any` | `__streams.highWaterMark` | 📋 Planned | - |
| `ignoreBOM` | `any` | `__streams.ignoreBOM` | 📋 Planned | - |
| `locked` | `any` | `__streams.locked` | 📋 Planned | - |
| `readable` | `any` | `__streams.readable` | 📋 Planned | - |
| `readableByteStreamController.close()` | `(...) => any` | `__streams.readableByteStreamController.close` | 📋 Planned | - |
| `readableByteStreamController.enqueue(chunk)` | `(...) => any` | `__streams.readableByteStreamController.enqueue` | 📋 Planned | - |
| `readableByteStreamController.error([error])` | `(...) => any` | `__streams.readableByteStreamController.error` | 📋 Planned | - |
| `readableStream.cancel([reason])` | `(...) => any` | `__streams.readableStream.cancel` | 📋 Planned | - |
| `readableStream.getReader([options])` | `(...) => any` | `__streams.readableStream.getReader` | 📋 Planned | - |
| `readableStream.pipeThrough(transform[, options])` | `(...) => any` | `__streams.readableStream.pipeThrough` | 📋 Planned | - |
| `readableStream.pipeTo(destination[, options])` | `(...) => any` | `__streams.readableStream.pipeTo` | 📋 Planned | - |
| `readableStream.tee()` | `(...) => any` | `__streams.readableStream.tee` | 📋 Planned | - |
| `readableStream.values([options])` | `(...) => any` | `__streams.readableStream.values` | 📋 Planned | - |
| `readableStreamBYOBReader.cancel([reason])` | `(...) => any` | `__streams.readableStreamBYOBReader.cancel` | 📋 Planned | - |
| `readableStreamBYOBReader.read(view[, options])` | `(...) => any` | `__streams.readableStreamBYOBReader.read` | 📋 Planned | - |
| `readableStreamBYOBReader.releaseLock()` | `(...) => any` | `__streams.readableStreamBYOBReader.releaseLock` | 📋 Planned | - |
| `readableStreamBYOBRequest.respond(bytesWritten)` | `(...) => any` | `__streams.readableStreamBYOBRequest.respond` | 📋 Planned | - |
| `readableStreamBYOBRequest.respondWithNewView(view)` | `(...) => any` | `__streams.readableStreamBYOBRequest.respondWithNewView` | 📋 Planned | - |
| `readableStreamDefaultController.close()` | `(...) => any` | `__streams.readableStreamDefaultController.close` | 📋 Planned | - |
| `readableStreamDefaultController.enqueue([chunk])` | `(...) => any` | `__streams.readableStreamDefaultController.enqueue` | 📋 Planned | - |
| `readableStreamDefaultController.error([error])` | `(...) => any` | `__streams.readableStreamDefaultController.error` | 📋 Planned | - |
| `readableStreamDefaultReader.cancel([reason])` | `(...) => any` | `__streams.readableStreamDefaultReader.cancel` | 📋 Planned | - |
| `readableStreamDefaultReader.read()` | `(...) => any` | `__streams.readableStreamDefaultReader.read` | 📋 Planned | - |
| `readableStreamDefaultReader.releaseLock()` | `(...) => any` | `__streams.readableStreamDefaultReader.releaseLock` | 📋 Planned | - |
| `ready` | `any` | `__streams.ready` | 📋 Planned | - |
| `signal` | `any` | `__streams.signal` | 📋 Planned | - |
| `size` | `any` | `__streams.size` | 📋 Planned | - |
| `streamConsumers.arrayBuffer(stream)` | `(...) => any` | `__streams.streamConsumers.arrayBuffer` | 📋 Planned | - |
| `streamConsumers.blob(stream)` | `(...) => any` | `__streams.streamConsumers.blob` | 📋 Planned | - |
| `streamConsumers.buffer(stream)` | `(...) => any` | `__streams.streamConsumers.buffer` | 📋 Planned | - |
| `streamConsumers.json(stream)` | `(...) => any` | `__streams.streamConsumers.json` | 📋 Planned | - |
| `streamConsumers.text(stream)` | `(...) => any` | `__streams.streamConsumers.text` | 📋 Planned | - |
| `transformStreamDefaultController.enqueue([chunk])` | `(...) => any` | `__streams.transformStreamDefaultController.enqueue` | 📋 Planned | - |
| `transformStreamDefaultController.error([reason])` | `(...) => any` | `__streams.transformStreamDefaultController.error` | 📋 Planned | - |
| `transformStreamDefaultController.terminate()` | `(...) => any` | `__streams.transformStreamDefaultController.terminate` | 📋 Planned | - |
| `view` | `any` | `__streams.view` | 📋 Planned | - |
| `writable` | `any` | `__streams.writable` | 📋 Planned | - |
| `writableStream.abort([reason])` | `(...) => any` | `__streams.writableStream.abort` | 📋 Planned | - |
| `writableStream.close()` | `(...) => any` | `__streams.writableStream.close` | 📋 Planned | - |
| `writableStream.getWriter()` | `(...) => any` | `__streams.writableStream.getWriter` | 📋 Planned | - |
| `writableStreamDefaultController.error([error])` | `(...) => any` | `__streams.writableStreamDefaultController.error` | 📋 Planned | - |
| `writableStreamDefaultWriter.abort([reason])` | `(...) => any` | `__streams.writableStreamDefaultWriter.abort` | 📋 Planned | - |
| `writableStreamDefaultWriter.close()` | `(...) => any` | `__streams.writableStreamDefaultWriter.close` | 📋 Planned | - |
| `writableStreamDefaultWriter.releaseLock()` | `(...) => any` | `__streams.writableStreamDefaultWriter.releaseLock` | 📋 Planned | - |
| `writableStreamDefaultWriter.write([chunk])` | `(...) => any` | `__streams.writableStreamDefaultWriter.write` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `streams` are organized per API under `internal/compiler/testdata/corpus/streams/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/streams/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
