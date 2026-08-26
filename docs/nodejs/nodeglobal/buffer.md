# Buffer Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Buffer Global Documentation](https://nodejs.org/docs/latest-v22.x/api/buffer.html)  
> **Type Definition Source**: [@types/node/buffer.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-buffer-*.js)

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
| `Blob` | `(...) => any` | `__buffer.Blob` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer` | `(...) => any` | `__buffer.Buffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `File` | `(...) => any` | `__buffer.File` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `INSPECT_MAX_BYTES` | `any` | `__buffer.INSPECT_MAX_BYTES` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `SlowBuffer` | `(...) => any` | `__buffer.SlowBuffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.arrayBuffer()` | `(...) => any` | `__buffer.blob.arrayBuffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.bytes()` | `(...) => any` | `__buffer.blob.bytes` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.size` | `any` | `__buffer.blob.size` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.slice([start[, end[, type]]])` | `(...) => any` | `__buffer.blob.slice` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.stream()` | `(...) => any` | `__buffer.blob.stream` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `blob.text()` | `(...) => any` | `__buffer.blob.text` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.compare(target[, targetStart[, targetEnd[, sourceStart[, sourceEnd]]]])` | `(...) => any` | `__buffer.buf.compare` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.copy(target[, targetStart[, sourceStart[, sourceEnd]]])` | `(...) => any` | `__buffer.buf.copy` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.entries()` | `(...) => any` | `__buffer.buf.entries` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.equals(otherBuffer)` | `(...) => any` | `__buffer.buf.equals` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.fill(value[, offset[, end]][, encoding])` | `(...) => any` | `__buffer.buf.fill` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.includes(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.includes` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.indexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.indexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.keys()` | `(...) => any` | `__buffer.buf.keys` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.lastIndexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__buffer.buf.lastIndexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.parent` | `any` | `__buffer.buf.parent` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readBigInt64BE([offset])` | `(...) => any` | `__buffer.buf.readBigInt64BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readBigInt64LE([offset])` | `(...) => any` | `__buffer.buf.readBigInt64LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readBigUInt64BE([offset])` | `(...) => any` | `__buffer.buf.readBigUInt64BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readBigUInt64LE([offset])` | `(...) => any` | `__buffer.buf.readBigUInt64LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readDoubleBE([offset])` | `(...) => any` | `__buffer.buf.readDoubleBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readDoubleLE([offset])` | `(...) => any` | `__buffer.buf.readDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readFloatBE([offset])` | `(...) => any` | `__buffer.buf.readFloatBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readFloatLE([offset])` | `(...) => any` | `__buffer.buf.readFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt16BE([offset])` | `(...) => any` | `__buffer.buf.readInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt16LE([offset])` | `(...) => any` | `__buffer.buf.readInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt32BE([offset])` | `(...) => any` | `__buffer.buf.readInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt32LE([offset])` | `(...) => any` | `__buffer.buf.readInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readInt8([offset])` | `(...) => any` | `__buffer.buf.readInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readIntBE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readIntBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readIntLE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readIntLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt16BE([offset])` | `(...) => any` | `__buffer.buf.readUInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt16LE([offset])` | `(...) => any` | `__buffer.buf.readUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt32BE([offset])` | `(...) => any` | `__buffer.buf.readUInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt32LE([offset])` | `(...) => any` | `__buffer.buf.readUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUInt8([offset])` | `(...) => any` | `__buffer.buf.readUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUIntBE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readUIntBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.readUIntLE(offset, byteLength)` | `(...) => any` | `__buffer.buf.readUIntLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.slice([start[, end]])` | `(...) => any` | `__buffer.buf.slice` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.subarray([start[, end]])` | `(...) => any` | `__buffer.buf.subarray` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.swap16()` | `(...) => any` | `__buffer.buf.swap16` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.swap32()` | `(...) => any` | `__buffer.buf.swap32` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.swap64()` | `(...) => any` | `__buffer.buf.swap64` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.toJSON()` | `(...) => any` | `__buffer.buf.toJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.toString([encoding[, start[, end]]])` | `(...) => any` | `__buffer.buf.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.values()` | `(...) => any` | `__buffer.buf.values` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.write(string[, offset[, length]][, encoding])` | `(...) => any` | `__buffer.buf.write` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeBigInt64BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigInt64BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeBigInt64LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigInt64LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeBigUInt64BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigUInt64BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeBigUInt64LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeBigUInt64LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeDoubleBE(value[, offset])` | `(...) => any` | `__buffer.buf.writeDoubleBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeDoubleLE(value[, offset])` | `(...) => any` | `__buffer.buf.writeDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeFloatBE(value[, offset])` | `(...) => any` | `__buffer.buf.writeFloatBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeFloatLE(value[, offset])` | `(...) => any` | `__buffer.buf.writeFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt16BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt16LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt32BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt32LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeInt8(value[, offset])` | `(...) => any` | `__buffer.buf.writeInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeIntBE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeIntBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeIntLE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeIntLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt16BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt16LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt32BE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt32LE(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUInt8(value[, offset])` | `(...) => any` | `__buffer.buf.writeUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUIntBE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeUIntBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buf.writeUIntLE(value, offset, byteLength)` | `(...) => any` | `__buffer.buf.writeUIntLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer` | `any` | `__buffer.buffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.atob(data)` | `(...) => any` | `__buffer.buffer.atob` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.btoa(data)` | `(...) => any` | `__buffer.buffer.btoa` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.isAscii(input)` | `(...) => any` | `__buffer.buffer.isAscii` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.isUtf8(input)` | `(...) => any` | `__buffer.buffer.isUtf8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.resolveObjectURL(id)` | `(...) => any` | `__buffer.buffer.resolveObjectURL` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `buffer.transcode(source, fromEnc, toEnc)` | `(...) => any` | `__buffer.buffer.transcode` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `byteOffset` | `any` | `__buffer.byteOffset` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `kMaxLength` | `any` | `__buffer.kMaxLength` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `kStringMaxLength` | `any` | `__buffer.kStringMaxLength` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `length` | `any` | `__buffer.length` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `poolSize` | `any` | `__buffer.poolSize` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `[index]` `index` {integer}` | `any` | `__buffer.[index]` `index` {integer}` | 📋 Planned | - |
| `lastModified` | `any` | `__buffer.lastModified` | 📋 Planned | - |
| `name` | `any` | `__buffer.name` | 📋 Planned | - |
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
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/buffer/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
