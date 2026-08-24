# Web Streams API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Web Streams API Documentation](https://nodejs.org/docs/latest-v22.x/api/web_streams_api.html)  
> **Type Definition Source**: [@types/node/web_streams_api.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-web_streams_api-*.js)

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
| `ByteLengthQueuingStrategy` | `(...) => any` | `__web_streams_api.ByteLengthQueuingStrategy` | 📋 Planned | - |
| `CompressionStream` | `(...) => any` | `__web_streams_api.CompressionStream` | 📋 Planned | - |
| `CountQueuingStrategy` | `(...) => any` | `__web_streams_api.CountQueuingStrategy` | 📋 Planned | - |
| `DecompressionStream` | `(...) => any` | `__web_streams_api.DecompressionStream` | 📋 Planned | - |
| `ReadableByteStreamController` | `(...) => any` | `__web_streams_api.ReadableByteStreamController` | 📋 Planned | - |
| `ReadableStream` | `(...) => any` | `__web_streams_api.ReadableStream` | 📋 Planned | - |
| `ReadableStream.from(iterable)` | `(...) => any` | `__web_streams_api.ReadableStream.from` | 📋 Planned | - |
| `ReadableStreamBYOBReader` | `(...) => any` | `__web_streams_api.ReadableStreamBYOBReader` | 📋 Planned | - |
| `ReadableStreamBYOBRequest` | `(...) => any` | `__web_streams_api.ReadableStreamBYOBRequest` | 📋 Planned | - |
| `ReadableStreamDefaultController` | `(...) => any` | `__web_streams_api.ReadableStreamDefaultController` | 📋 Planned | - |
| `ReadableStreamDefaultReader` | `(...) => any` | `__web_streams_api.ReadableStreamDefaultReader` | 📋 Planned | - |
| `TextDecoderStream` | `(...) => any` | `__web_streams_api.TextDecoderStream` | 📋 Planned | - |
| `TextEncoderStream` | `(...) => any` | `__web_streams_api.TextEncoderStream` | 📋 Planned | - |
| `TransformStream` | `(...) => any` | `__web_streams_api.TransformStream` | 📋 Planned | - |
| `TransformStreamDefaultController` | `(...) => any` | `__web_streams_api.TransformStreamDefaultController` | 📋 Planned | - |
| `WritableStream` | `(...) => any` | `__web_streams_api.WritableStream` | 📋 Planned | - |
| `WritableStreamDefaultController` | `(...) => any` | `__web_streams_api.WritableStreamDefaultController` | 📋 Planned | - |
| `WritableStreamDefaultWriter` | `(...) => any` | `__web_streams_api.WritableStreamDefaultWriter` | 📋 Planned | - |
| `byobRequest` | `any` | `__web_streams_api.byobRequest` | 📋 Planned | - |
| `closed` | `any` | `__web_streams_api.closed` | 📋 Planned | - |
| `desiredSize` | `any` | `__web_streams_api.desiredSize` | 📋 Planned | - |
| `encoding` | `any` | `__web_streams_api.encoding` | 📋 Planned | - |
| `fatal` | `any` | `__web_streams_api.fatal` | 📋 Planned | - |
| `highWaterMark` | `any` | `__web_streams_api.highWaterMark` | 📋 Planned | - |
| `ignoreBOM` | `any` | `__web_streams_api.ignoreBOM` | 📋 Planned | - |
| `locked` | `any` | `__web_streams_api.locked` | 📋 Planned | - |
| `readable` | `any` | `__web_streams_api.readable` | 📋 Planned | - |
| `readableByteStreamController.close()` | `(...) => any` | `__web_streams_api.readableByteStreamController.close` | 📋 Planned | - |
| `readableByteStreamController.enqueue(chunk)` | `(...) => any` | `__web_streams_api.readableByteStreamController.enqueue` | 📋 Planned | - |
| `readableByteStreamController.error([error])` | `(...) => any` | `__web_streams_api.readableByteStreamController.error` | 📋 Planned | - |
| `readableStream.cancel([reason])` | `(...) => any` | `__web_streams_api.readableStream.cancel` | 📋 Planned | - |
| `readableStream.getReader([options])` | `(...) => any` | `__web_streams_api.readableStream.getReader` | 📋 Planned | - |
| `readableStream.pipeThrough(transform[, options])` | `(...) => any` | `__web_streams_api.readableStream.pipeThrough` | 📋 Planned | - |
| `readableStream.pipeTo(destination[, options])` | `(...) => any` | `__web_streams_api.readableStream.pipeTo` | 📋 Planned | - |
| `readableStream.tee()` | `(...) => any` | `__web_streams_api.readableStream.tee` | 📋 Planned | - |
| `readableStream.values([options])` | `(...) => any` | `__web_streams_api.readableStream.values` | 📋 Planned | - |
| `readableStreamBYOBReader.cancel([reason])` | `(...) => any` | `__web_streams_api.readableStreamBYOBReader.cancel` | 📋 Planned | - |
| `readableStreamBYOBReader.read(view[, options])` | `(...) => any` | `__web_streams_api.readableStreamBYOBReader.read` | 📋 Planned | - |
| `readableStreamBYOBReader.releaseLock()` | `(...) => any` | `__web_streams_api.readableStreamBYOBReader.releaseLock` | 📋 Planned | - |
| `readableStreamBYOBRequest.respond(bytesWritten)` | `(...) => any` | `__web_streams_api.readableStreamBYOBRequest.respond` | 📋 Planned | - |
| `readableStreamBYOBRequest.respondWithNewView(view)` | `(...) => any` | `__web_streams_api.readableStreamBYOBRequest.respondWithNewView` | 📋 Planned | - |
| `readableStreamDefaultController.close()` | `(...) => any` | `__web_streams_api.readableStreamDefaultController.close` | 📋 Planned | - |
| `readableStreamDefaultController.enqueue([chunk])` | `(...) => any` | `__web_streams_api.readableStreamDefaultController.enqueue` | 📋 Planned | - |
| `readableStreamDefaultController.error([error])` | `(...) => any` | `__web_streams_api.readableStreamDefaultController.error` | 📋 Planned | - |
| `readableStreamDefaultReader.cancel([reason])` | `(...) => any` | `__web_streams_api.readableStreamDefaultReader.cancel` | 📋 Planned | - |
| `readableStreamDefaultReader.read()` | `(...) => any` | `__web_streams_api.readableStreamDefaultReader.read` | 📋 Planned | - |
| `readableStreamDefaultReader.releaseLock()` | `(...) => any` | `__web_streams_api.readableStreamDefaultReader.releaseLock` | 📋 Planned | - |
| `ready` | `any` | `__web_streams_api.ready` | 📋 Planned | - |
| `signal` | `any` | `__web_streams_api.signal` | 📋 Planned | - |
| `size` | `any` | `__web_streams_api.size` | 📋 Planned | - |
| `transformStreamDefaultController.enqueue([chunk])` | `(...) => any` | `__web_streams_api.transformStreamDefaultController.enqueue` | 📋 Planned | - |
| `transformStreamDefaultController.error([reason])` | `(...) => any` | `__web_streams_api.transformStreamDefaultController.error` | 📋 Planned | - |
| `transformStreamDefaultController.terminate()` | `(...) => any` | `__web_streams_api.transformStreamDefaultController.terminate` | 📋 Planned | - |
| `view` | `any` | `__web_streams_api.view` | 📋 Planned | - |
| `writable` | `any` | `__web_streams_api.writable` | 📋 Planned | - |
| `writableStream.abort([reason])` | `(...) => any` | `__web_streams_api.writableStream.abort` | 📋 Planned | - |
| `writableStream.close()` | `(...) => any` | `__web_streams_api.writableStream.close` | 📋 Planned | - |
| `writableStream.getWriter()` | `(...) => any` | `__web_streams_api.writableStream.getWriter` | 📋 Planned | - |
| `writableStreamDefaultController.error([error])` | `(...) => any` | `__web_streams_api.writableStreamDefaultController.error` | 📋 Planned | - |
| `writableStreamDefaultWriter.abort([reason])` | `(...) => any` | `__web_streams_api.writableStreamDefaultWriter.abort` | 📋 Planned | - |
| `writableStreamDefaultWriter.close()` | `(...) => any` | `__web_streams_api.writableStreamDefaultWriter.close` | 📋 Planned | - |
| `writableStreamDefaultWriter.releaseLock()` | `(...) => any` | `__web_streams_api.writableStreamDefaultWriter.releaseLock` | 📋 Planned | - |
| `writableStreamDefaultWriter.write([chunk])` | `(...) => any` | `__web_streams_api.writableStreamDefaultWriter.write` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `web_streams_api` are organized per API under `internal/compiler/testdata/corpus/web_streams_api/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/web_streams_api/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
