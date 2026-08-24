# Stream Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:stream`  
> **Specification Reference**: [Node.js 22 LTS Stream Documentation](https://nodejs.org/docs/latest-v22.x/api/stream.html)  
> **Type Definition Source**: [@types/node/stream.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-stream-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:stream`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `readable.push('')` | `(...) => any` | `__stream.readable.push` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `readable.read(0)` | `(...) => any` | `__stream.readable.read` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Duplex.from(src)` | `(...) => any` | `__stream.stream.Duplex.from` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Duplex.fromWeb(pair[, options])` | `(...) => any` | `__stream.stream.Duplex.fromWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Duplex.toWeb(streamDuplex)` | `(...) => any` | `__stream.stream.Duplex.toWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Readable.from(iterable[, options])` | `(...) => any` | `__stream.stream.Readable.from` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Readable.fromWeb(readableStream[, options])` | `(...) => any` | `__stream.stream.Readable.fromWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Readable.isDisturbed(stream)` | `(...) => any` | `__stream.stream.Readable.isDisturbed` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Readable.toWeb(streamReadable[, options])` | `(...) => any` | `__stream.stream.Readable.toWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Writable.fromWeb(writableStream[, options])` | `(...) => any` | `__stream.stream.Writable.fromWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.Writable.toWeb(streamWritable)` | `(...) => any` | `__stream.stream.Writable.toWeb` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.addAbortSignal(signal, stream)` | `(...) => any` | `__stream.stream.addAbortSignal` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.compose(...streams)` | `(...) => any` | `__stream.stream.compose` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.finished(stream[, options])` | `(...) => any` | `__stream.stream.finished` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.finished(stream[, options], callback)` | `(...) => any` | `__stream.stream.finished` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.getDefaultHighWaterMark(objectMode)` | `(...) => any` | `__stream.stream.getDefaultHighWaterMark` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.isErrored(stream)` | `(...) => any` | `__stream.stream.isErrored` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.isReadable(stream)` | `(...) => any` | `__stream.stream.isReadable` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.isWritable(stream)` | `(...) => any` | `__stream.stream.isWritable` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.pipeline(source[, ...transforms], destination, callback)` | `(...) => any` | `__stream.stream.pipeline` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.pipeline(source[, ...transforms], destination[, options])` | `(...) => any` | `__stream.stream.pipeline` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.pipeline(streams, callback)` | `(...) => any` | `__stream.stream.pipeline` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.pipeline(streams[, options])` | `(...) => any` | `__stream.stream.pipeline` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |
| `stream.setDefaultHighWaterMark(objectMode, value)` | `(...) => any` | `__stream.stream.setDefaultHighWaterMark` | ✅ Done | `internal/compiler/testdata/corpus/api/stream.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `stream` are organized per API under `internal/compiler/testdata/corpus/stream/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/stream/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
