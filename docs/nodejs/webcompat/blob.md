# WHATWG Blob API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG WHATWG Blob API Specification](https://wintercg.org/)  
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
| `Buffer` | `(...) => any` | `__blob.Buffer` | ✅ Done | `internal/compiler/testdata/corpus/buffer/is_buffer/` |
| `buffer` | `any` | `__blob.buffer` | ✅ Done | `internal/compiler/testdata/corpus/buffer/is_buffer/` |
| `Blob` | `(...) => any` | `__blob.Blob` | 📋 Planned | - |
| `File` | `(...) => any` | `__blob.File` | 📋 Planned | - |
| `INSPECT_MAX_BYTES` | `any` | `__blob.INSPECT_MAX_BYTES` | 📋 Planned | - |
| `MAX_LENGTH` | `any` | `__blob.MAX_LENGTH` | 📋 Planned | - |
| `MAX_STRING_LENGTH` | `any` | `__blob.MAX_STRING_LENGTH` | 📋 Planned | - |
| `SlowBuffer` | `(...) => any` | `__blob.SlowBuffer` | 📋 Planned | - |
| `[index]` `index` {integer}` | `any` | `__blob.[index]` `index` {integer}` | 📋 Planned | - |
| `blob.arrayBuffer()` | `(...) => any` | `__blob.blob.arrayBuffer` | 📋 Planned | - |
| `blob.bytes()` | `(...) => any` | `__blob.blob.bytes` | 📋 Planned | - |
| `blob.size` | `any` | `__blob.blob.size` | 📋 Planned | - |
| `blob.slice([start[, end[, type]]])` | `(...) => any` | `__blob.blob.slice` | 📋 Planned | - |
| `blob.stream()` | `(...) => any` | `__blob.blob.stream` | 📋 Planned | - |
| `blob.text()` | `(...) => any` | `__blob.blob.text` | 📋 Planned | - |
| `buf.compare(target[, targetStart[, targetEnd[, sourceStart[, sourceEnd]]]])` | `(...) => any` | `__blob.buf.compare` | 📋 Planned | - |
| `buf.copy(target[, targetStart[, sourceStart[, sourceEnd]]])` | `(...) => any` | `__blob.buf.copy` | 📋 Planned | - |
| `buf.entries()` | `(...) => any` | `__blob.buf.entries` | 📋 Planned | - |
| `buf.equals(otherBuffer)` | `(...) => any` | `__blob.buf.equals` | 📋 Planned | - |
| `buf.fill(value[, offset[, end]][, encoding])` | `(...) => any` | `__blob.buf.fill` | 📋 Planned | - |
| `buf.includes(value[, byteOffset][, encoding])` | `(...) => any` | `__blob.buf.includes` | 📋 Planned | - |
| `buf.indexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__blob.buf.indexOf` | 📋 Planned | - |
| `buf.keys()` | `(...) => any` | `__blob.buf.keys` | 📋 Planned | - |
| `buf.lastIndexOf(value[, byteOffset][, encoding])` | `(...) => any` | `__blob.buf.lastIndexOf` | 📋 Planned | - |
| `buf.parent` | `any` | `__blob.buf.parent` | 📋 Planned | - |
| `buf.readBigInt64BE([offset])` | `(...) => any` | `__blob.buf.readBigInt64BE` | 📋 Planned | - |
| `buf.readBigInt64LE([offset])` | `(...) => any` | `__blob.buf.readBigInt64LE` | 📋 Planned | - |
| `buf.readBigUInt64BE([offset])` | `(...) => any` | `__blob.buf.readBigUInt64BE` | 📋 Planned | - |
| `buf.readBigUInt64LE([offset])` | `(...) => any` | `__blob.buf.readBigUInt64LE` | 📋 Planned | - |
| `buf.readDoubleBE([offset])` | `(...) => any` | `__blob.buf.readDoubleBE` | 📋 Planned | - |
| `buf.readDoubleLE([offset])` | `(...) => any` | `__blob.buf.readDoubleLE` | 📋 Planned | - |
| `buf.readFloatBE([offset])` | `(...) => any` | `__blob.buf.readFloatBE` | 📋 Planned | - |
| `buf.readFloatLE([offset])` | `(...) => any` | `__blob.buf.readFloatLE` | 📋 Planned | - |
| `buf.readInt16BE([offset])` | `(...) => any` | `__blob.buf.readInt16BE` | 📋 Planned | - |
| `buf.readInt16LE([offset])` | `(...) => any` | `__blob.buf.readInt16LE` | 📋 Planned | - |
| `buf.readInt32BE([offset])` | `(...) => any` | `__blob.buf.readInt32BE` | 📋 Planned | - |
| `buf.readInt32LE([offset])` | `(...) => any` | `__blob.buf.readInt32LE` | 📋 Planned | - |
| `buf.readInt8([offset])` | `(...) => any` | `__blob.buf.readInt8` | 📋 Planned | - |
| `buf.readIntBE(offset, byteLength)` | `(...) => any` | `__blob.buf.readIntBE` | 📋 Planned | - |
| `buf.readIntLE(offset, byteLength)` | `(...) => any` | `__blob.buf.readIntLE` | 📋 Planned | - |
| `buf.readUInt16BE([offset])` | `(...) => any` | `__blob.buf.readUInt16BE` | 📋 Planned | - |
| `buf.readUInt16LE([offset])` | `(...) => any` | `__blob.buf.readUInt16LE` | 📋 Planned | - |
| `buf.readUInt32BE([offset])` | `(...) => any` | `__blob.buf.readUInt32BE` | 📋 Planned | - |
| `buf.readUInt32LE([offset])` | `(...) => any` | `__blob.buf.readUInt32LE` | 📋 Planned | - |
| `buf.readUInt8([offset])` | `(...) => any` | `__blob.buf.readUInt8` | 📋 Planned | - |
| `buf.readUIntBE(offset, byteLength)` | `(...) => any` | `__blob.buf.readUIntBE` | 📋 Planned | - |
| `buf.readUIntLE(offset, byteLength)` | `(...) => any` | `__blob.buf.readUIntLE` | 📋 Planned | - |
| `buf.slice([start[, end]])` | `(...) => any` | `__blob.buf.slice` | 📋 Planned | - |
| `buf.subarray([start[, end]])` | `(...) => any` | `__blob.buf.subarray` | 📋 Planned | - |
| `buf.swap16()` | `(...) => any` | `__blob.buf.swap16` | 📋 Planned | - |
| `buf.swap32()` | `(...) => any` | `__blob.buf.swap32` | 📋 Planned | - |
| `buf.swap64()` | `(...) => any` | `__blob.buf.swap64` | 📋 Planned | - |
| `buf.toJSON()` | `(...) => any` | `__blob.buf.toJSON` | 📋 Planned | - |
| `buf.toString([encoding[, start[, end]]])` | `(...) => any` | `__blob.buf.toString` | 📋 Planned | - |
| `buf.values()` | `(...) => any` | `__blob.buf.values` | 📋 Planned | - |
| `buf.write(string[, offset[, length]][, encoding])` | `(...) => any` | `__blob.buf.write` | 📋 Planned | - |
| `buf.writeBigInt64BE(value[, offset])` | `(...) => any` | `__blob.buf.writeBigInt64BE` | 📋 Planned | - |
| `buf.writeBigInt64LE(value[, offset])` | `(...) => any` | `__blob.buf.writeBigInt64LE` | 📋 Planned | - |
| `buf.writeBigUInt64BE(value[, offset])` | `(...) => any` | `__blob.buf.writeBigUInt64BE` | 📋 Planned | - |
| `buf.writeBigUInt64LE(value[, offset])` | `(...) => any` | `__blob.buf.writeBigUInt64LE` | 📋 Planned | - |
| `buf.writeDoubleBE(value[, offset])` | `(...) => any` | `__blob.buf.writeDoubleBE` | 📋 Planned | - |
| `buf.writeDoubleLE(value[, offset])` | `(...) => any` | `__blob.buf.writeDoubleLE` | 📋 Planned | - |
| `buf.writeFloatBE(value[, offset])` | `(...) => any` | `__blob.buf.writeFloatBE` | 📋 Planned | - |
| `buf.writeFloatLE(value[, offset])` | `(...) => any` | `__blob.buf.writeFloatLE` | 📋 Planned | - |
| `buf.writeInt16BE(value[, offset])` | `(...) => any` | `__blob.buf.writeInt16BE` | 📋 Planned | - |
| `buf.writeInt16LE(value[, offset])` | `(...) => any` | `__blob.buf.writeInt16LE` | 📋 Planned | - |
| `buf.writeInt32BE(value[, offset])` | `(...) => any` | `__blob.buf.writeInt32BE` | 📋 Planned | - |
| `buf.writeInt32LE(value[, offset])` | `(...) => any` | `__blob.buf.writeInt32LE` | 📋 Planned | - |
| `buf.writeInt8(value[, offset])` | `(...) => any` | `__blob.buf.writeInt8` | 📋 Planned | - |
| `buf.writeIntBE(value, offset, byteLength)` | `(...) => any` | `__blob.buf.writeIntBE` | 📋 Planned | - |
| `buf.writeIntLE(value, offset, byteLength)` | `(...) => any` | `__blob.buf.writeIntLE` | 📋 Planned | - |
| `buf.writeUInt16BE(value[, offset])` | `(...) => any` | `__blob.buf.writeUInt16BE` | 📋 Planned | - |
| `buf.writeUInt16LE(value[, offset])` | `(...) => any` | `__blob.buf.writeUInt16LE` | 📋 Planned | - |
| `buf.writeUInt32BE(value[, offset])` | `(...) => any` | `__blob.buf.writeUInt32BE` | 📋 Planned | - |
| `buf.writeUInt32LE(value[, offset])` | `(...) => any` | `__blob.buf.writeUInt32LE` | 📋 Planned | - |
| `buf.writeUInt8(value[, offset])` | `(...) => any` | `__blob.buf.writeUInt8` | 📋 Planned | - |
| `buf.writeUIntBE(value, offset, byteLength)` | `(...) => any` | `__blob.buf.writeUIntBE` | 📋 Planned | - |
| `buf.writeUIntLE(value, offset, byteLength)` | `(...) => any` | `__blob.buf.writeUIntLE` | 📋 Planned | - |
| `buffer.atob(data)` | `(...) => any` | `__blob.buffer.atob` | 📋 Planned | - |
| `buffer.btoa(data)` | `(...) => any` | `__blob.buffer.btoa` | 📋 Planned | - |
| `buffer.isAscii(input)` | `(...) => any` | `__blob.buffer.isAscii` | 📋 Planned | - |
| `buffer.isUtf8(input)` | `(...) => any` | `__blob.buffer.isUtf8` | 📋 Planned | - |
| `buffer.resolveObjectURL(id)` | `(...) => any` | `__blob.buffer.resolveObjectURL` | 📋 Planned | - |
| `buffer.transcode(source, fromEnc, toEnc)` | `(...) => any` | `__blob.buffer.transcode` | 📋 Planned | - |
| `byteOffset` | `any` | `__blob.byteOffset` | 📋 Planned | - |
| `kMaxLength` | `any` | `__blob.kMaxLength` | 📋 Planned | - |
| `kStringMaxLength` | `any` | `__blob.kStringMaxLength` | 📋 Planned | - |
| `lastModified` | `any` | `__blob.lastModified` | 📋 Planned | - |
| `length` | `any` | `__blob.length` | 📋 Planned | - |
| `name` | `any` | `__blob.name` | 📋 Planned | - |
| `poolSize` | `any` | `__blob.poolSize` | 📋 Planned | - |
| `type` | `any` | `__blob.type` | 📋 Planned | - |

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
