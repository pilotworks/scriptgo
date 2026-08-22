# Buffer Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:buffer`  
> **Specification Reference**: [Node.js 22 LTS Buffer Documentation](https://nodejs.org/docs/latest-v22.x/api/buffer.html)  
> **Type Definition Source**: [@types/node/buffer.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-buffer-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:buffer`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `buf.compare(target[, targetStart[, targetEnd[, sourceStart[, sourceEnd]]]])` | `(...) => any` | `__buffer.buf.compare` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.copy(target[, targetStart[, sourceStart[, sourceEnd]]])` | `(...) => any` | `__buffer.buf.copy` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.equals(otherBuffer)` | `(...) => any` | `__buffer.buf.equals` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.indexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.indexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readDoubleLE([offset])` | `(...) => any` | `__buffer.buf.readDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readFloatLE([offset])` | `(...) => any` | `__buffer.buf.readFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt32LE([offset])` | `(...) => any` | `__buffer.buf.readInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt16LE([offset])` | `(...) => any` | `__buffer.buf.readUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt32LE([offset])` | `(...) => any` | `__buffer.buf.readUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt8([offset])` | `(...) => any` | `__buffer.buf.readUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.toString([encoding[, start[, end]]])` | `(...) => any` | `__buffer.buf.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeDoubleLE(value[, offset])` | `(...) => any` | `__buffer.buf.writeDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeFloatLE(value[, offset])` | `(...) => any` | `__buffer.buf.writeFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt32LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt16LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt32LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt8(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Blob` | `(...) => any` | `__buffer.Blob` | 📋 Planned | - |
| `Buffer` | `(...) => any` | `__buffer.Buffer` | 📋 Planned | - |
| `File` | `(...) => any` | `__buffer.File` | 📋 Planned | - |
| `INSPECT_MAX_BYTES` | `any` | `__buffer.INSPECT_MAX_BYTES` | 📋 Planned | - |
| `SlowBuffer` | `(...) => any` | `__buffer.SlowBuffer` | 📋 Planned | - |
| `[index]` `index` {integer}` | `any` | `__buffer.[index]` `index` {integer}` | 📋 Planned | - |
| `blob.arrayBuffer()` | `(...) => any` | `__buffer.blob.arrayBuffer` | 📋 Planned | - |
| `blob.bytes()` | `(...) => any` | `__buffer.blob.bytes` | 📋 Planned | - |
| `blob.size` | `any` | `__buffer.blob.size` | 📋 Planned | - |
| `blob.slice([start[, end[, type]]])` | `(...) => any` | `__buffer.blob.slice` | 📋 Planned | - |
| `blob.stream()` | `(...) => any` | `__buffer.blob.stream` | 📋 Planned | - |
| `blob.text()` | `(...) => any` | `__buffer.blob.text` | 📋 Planned | - |
| `buf.entries()` | `(...) => any` | `__buffer.buf.entries` | 📋 Planned | - |
| `buf.fill(value[, offset[, end]][, encoding])` | `(...) => any` | `__buffer.buf.fill` | 📋 Planned | - |
| `buf.includes(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.includes` | 📋 Planned | - |
| `buf.keys()` | `(...) => any` | `__buffer.buf.keys` | 📋 Planned | - |
| `buf.lastIndexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.lastIndexOf` | 📋 Planned | - |
| `buf.parent` | `any` | `__buffer.buf.parent` | 📋 Planned | - |
| `buf.readBigInt64BE([offset])` | `(...) => any` | `__buffer.buf.readBigInt64BE` | 📋 Planned | - |
| `buf.readBigInt64LE([offset])` | `(...) => any` | `__buffer.buf.readBigInt64LE` | 📋 Planned | - |
| `buf.readBigUInt64BE([offset])` | `(...) => any` | `__buffer.buf.readBigUInt64BE` | 📋 Planned | - |
| `buf.readBigUInt64LE([offset])` | `(...) => any` | `__buffer.buf.readBigUInt64LE` | 📋 Planned | - |
| `buf.readDoubleBE([offset])` | `(...) => any` | `__buffer.buf.readDoubleBE` | 📋 Planned | - |
| `buf.readFloatBE([offset])` | `(...) => any` | `__buffer.buf.readFloatBE` | 📋 Planned | - |
| `buf.readInt16BE([offset])` | `(...) => any` | `__buffer.buf.readInt16BE` | 📋 Planned | - |
| `buf.readInt16LE([offset])` | `(...) => any` | `__buffer.buf.readInt16LE` | 📋 Planned | - |
| `buf.readInt32BE([offset])` | `(...) => any` | `__buffer.buf.readInt32BE` | 📋 Planned | - |
| `buf.readInt8([offset])` | `(...) => any` | `__buffer.buf.readInt8` | 📋 Planned | - |
| `buf.readIntBE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readIntBE` | 📋 Planned | - |
| `buf.readIntLE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readIntLE` | 📋 Planned | - |
| `buf.readUInt16BE([offset])` | `(...) => any` | `__buffer.buf.readUInt16BE` | 📋 Planned | - |
| `buf.readUInt32BE([offset])` | `(...) => any` | `__buffer.buf.readUInt32BE` | 📋 Planned | - |
| `buf.readUIntBE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readUIntBE` | 📋 Planned | - |
| `buf.readUIntLE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readUIntLE` | 📋 Planned | - |
| `buf.slice([start[, end]])` | `(...) => any` | `__buffer.buf.slice` | 📋 Planned | - |
| `buf.subarray([start[, end]])` | `(...) => any` | `__buffer.buf.subarray` | 📋 Planned | - |
| `buf.swap16()` | `(...) => any` | `__buffer.buf.swap16` | 📋 Planned | - |
| `buf.swap32()` | `(...) => any` | `__buffer.buf.swap32` | 📋 Planned | - |
| `buf.swap64()` | `(...) => any` | `__buffer.buf.swap64` | 📋 Planned | - |
| `buf.toJSON()` | `(...) => any` | `__buffer.buf.toJSON` | 📋 Planned | - |
| `buf.values()` | `(...) => any` | `__buffer.buf.values` | 📋 Planned | - |
| `buf.write(string[, offset[, length]][, encoding])` | `(...) => any` | `__buffer.buf.write` | 📋 Planned | - |
| `buf.writeBigInt64BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigInt64BE` | 📋 Planned | - |
| `buf.writeBigInt64LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigInt64LE` | 📋 Planned | - |
| `buf.writeBigUInt64BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigUInt64BE` | 📋 Planned | - |
| `buf.writeBigUInt64LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigUInt64LE` | 📋 Planned | - |
| `buf.writeDoubleBE(value[, offset])` | `(...) => any` | `__buffer.buf.writeDoubleBE` | 📋 Planned | - |
| `buf.writeFloatBE(value[, offset])` | `(...) => any` | `__buffer.buf.writeFloatBE` | 📋 Planned | - |
| `buf.writeInt16BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt16BE` | 📋 Planned | - |
| `buf.writeInt16LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt16LE` | 📋 Planned | - |
| `buf.writeInt32BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt32BE` | 📋 Planned | - |
| `buf.writeInt8(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt8` | 📋 Planned | - |
| `buf.writeIntBE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeIntBE` | 📋 Planned | - |
| `buf.writeIntLE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeIntLE` | 📋 Planned | - |
| `buf.writeUInt16BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt16BE` | 📋 Planned | - |
| `buf.writeUInt32BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt32BE` | 📋 Planned | - |
| `buf.writeUIntBE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeUIntBE` | 📋 Planned | - |
| `buf.writeUIntLE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeUIntLE` | 📋 Planned | - |
| `buffer` | `any` | `__buffer.buffer` | 📋 Planned | - |
| `buffer.atob(data)` | `(...) => any` | `__buffer.buffer.atob` | 📋 Planned | - |
| `buffer.btoa(data)` | `(...) => any` | `__buffer.buffer.btoa` | 📋 Planned | - |
| `buffer.isAscii(input)` | `(...) => any` | `__buffer.buffer.isAscii` | 📋 Planned | - |
| `buffer.isUtf8(input)` | `(...) => any` | `__buffer.buffer.isUtf8` | 📋 Planned | - |
| `buffer.resolveObjectURL(id)` | `(...) => any` | `__buffer.buffer.resolveObjectURL` | 📋 Planned | - |
| `buffer.transcode(source, fromEnc, toEnc)` | `(...) => any` | `__buffer.buffer.transcode` | 📋 Planned | - |
| `byteOffset` | `any` | `__buffer.byteOffset` | 📋 Planned | - |
| `kMaxLength` | `any` | `__buffer.kMaxLength` | 📋 Planned | - |
| `kStringMaxLength` | `any` | `__buffer.kStringMaxLength` | 📋 Planned | - |
| `lastModified` | `any` | `__buffer.lastModified` | 📋 Planned | - |
| `length` | `any` | `__buffer.length` | 📋 Planned | - |
| `name` | `any` | `__buffer.name` | 📋 Planned | - |
| `poolSize` | `any` | `__buffer.poolSize` | 📋 Planned | - |
| `type` | `any` | `__buffer.type` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `buffer` are organized per API under `internal/compiler/testdata/corpus/buffer/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/buffer/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
