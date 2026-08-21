# Zlib Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:zlib`  
> **Specification Reference**: [Node.js 22 LTS Zlib Documentation](https://nodejs.org/docs/latest-v22.x/api/zlib.html)  
> **Type Definition Source**: [@types/node/zlib.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-zlib-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:zlib`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `bytesRead` {number}` | `any` | `__zlib.bytesRead` {number}` | 📋 Planned | - |
| `bytesWritten` | `any` | `__zlib.bytesWritten` | 📋 Planned | - |
| `zlib.BrotliCompress` | `(...) => any` | `__zlib.zlib.BrotliCompress` | 📋 Planned | - |
| `zlib.BrotliDecompress` | `(...) => any` | `__zlib.zlib.BrotliDecompress` | 📋 Planned | - |
| `zlib.Deflate` | `(...) => any` | `__zlib.zlib.Deflate` | 📋 Planned | - |
| `zlib.DeflateRaw` | `(...) => any` | `__zlib.zlib.DeflateRaw` | 📋 Planned | - |
| `zlib.Gunzip` | `(...) => any` | `__zlib.zlib.Gunzip` | 📋 Planned | - |
| `zlib.Gzip` | `(...) => any` | `__zlib.zlib.Gzip` | 📋 Planned | - |
| `zlib.Inflate` | `(...) => any` | `__zlib.zlib.Inflate` | 📋 Planned | - |
| `zlib.InflateRaw` | `(...) => any` | `__zlib.zlib.InflateRaw` | 📋 Planned | - |
| `zlib.Unzip` | `(...) => any` | `__zlib.zlib.Unzip` | 📋 Planned | - |
| `zlib.ZlibBase` | `(...) => any` | `__zlib.zlib.ZlibBase` | 📋 Planned | - |
| `zlib.ZstdCompress` | `(...) => any` | `__zlib.zlib.ZstdCompress` | 📋 Planned | - |
| `zlib.ZstdDecompress` | `(...) => any` | `__zlib.zlib.ZstdDecompress` | 📋 Planned | - |
| `zlib.brotliCompress(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.brotliCompress` | 📋 Planned | - |
| `zlib.brotliCompressSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.brotliCompressSync` | 📋 Planned | - |
| `zlib.brotliDecompress(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.brotliDecompress` | 📋 Planned | - |
| `zlib.brotliDecompressSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.brotliDecompressSync` | 📋 Planned | - |
| `zlib.close([callback])` | `(...) => any` | `__zlib.zlib.close` | 📋 Planned | - |
| `zlib.constants` | `any` | `__zlib.zlib.constants` | 📋 Planned | - |
| `zlib.crc32(data[, value])` | `(...) => any` | `__zlib.zlib.crc32` | 📋 Planned | - |
| `zlib.createBrotliCompress([options])` | `(...) => any` | `__zlib.zlib.createBrotliCompress` | 📋 Planned | - |
| `zlib.createBrotliDecompress([options])` | `(...) => any` | `__zlib.zlib.createBrotliDecompress` | 📋 Planned | - |
| `zlib.createDeflate([options])` | `(...) => any` | `__zlib.zlib.createDeflate` | 📋 Planned | - |
| `zlib.createDeflateRaw([options])` | `(...) => any` | `__zlib.zlib.createDeflateRaw` | 📋 Planned | - |
| `zlib.createGunzip([options])` | `(...) => any` | `__zlib.zlib.createGunzip` | 📋 Planned | - |
| `zlib.createGzip([options])` | `(...) => any` | `__zlib.zlib.createGzip` | 📋 Planned | - |
| `zlib.createInflate([options])` | `(...) => any` | `__zlib.zlib.createInflate` | 📋 Planned | - |
| `zlib.createInflateRaw([options])` | `(...) => any` | `__zlib.zlib.createInflateRaw` | 📋 Planned | - |
| `zlib.createUnzip([options])` | `(...) => any` | `__zlib.zlib.createUnzip` | 📋 Planned | - |
| `zlib.createZstdCompress([options])` | `(...) => any` | `__zlib.zlib.createZstdCompress` | 📋 Planned | - |
| `zlib.createZstdDecompress([options])` | `(...) => any` | `__zlib.zlib.createZstdDecompress` | 📋 Planned | - |
| `zlib.deflate(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.deflate` | 📋 Planned | - |
| `zlib.deflateRaw(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.deflateRaw` | 📋 Planned | - |
| `zlib.deflateRawSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.deflateRawSync` | 📋 Planned | - |
| `zlib.deflateSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.deflateSync` | 📋 Planned | - |
| `zlib.flush([kind, ]callback)` | `(...) => any` | `__zlib.zlib.flush` | 📋 Planned | - |
| `zlib.gunzip(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.gunzip` | 📋 Planned | - |
| `zlib.gunzipSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.gunzipSync` | 📋 Planned | - |
| `zlib.gzip(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.gzip` | 📋 Planned | - |
| `zlib.gzipSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.gzipSync` | 📋 Planned | - |
| `zlib.inflate(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.inflate` | 📋 Planned | - |
| `zlib.inflateRaw(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.inflateRaw` | 📋 Planned | - |
| `zlib.inflateRawSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.inflateRawSync` | 📋 Planned | - |
| `zlib.inflateSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.inflateSync` | 📋 Planned | - |
| `zlib.params(level, strategy, callback)` | `(...) => any` | `__zlib.zlib.params` | 📋 Planned | - |
| `zlib.reset()` | `(...) => any` | `__zlib.zlib.reset` | 📋 Planned | - |
| `zlib.unzip(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.unzip` | 📋 Planned | - |
| `zlib.unzipSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.unzipSync` | 📋 Planned | - |
| `zlib.zstdCompress(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.zstdCompress` | 📋 Planned | - |
| `zlib.zstdCompressSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.zstdCompressSync` | 📋 Planned | - |
| `zlib.zstdDecompress(buffer[, options], callback)` | `(...) => any` | `__zlib.zlib.zstdDecompress` | 📋 Planned | - |
| `zlib.zstdDecompressSync(buffer[, options])` | `(...) => any` | `__zlib.zlib.zstdDecompressSync` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `zlib` are organized per API under `internal/compiler/testdata/corpus/zlib/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/zlib/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
