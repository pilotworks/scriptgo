# Buffer Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Buffer Specification](https://tc39.es/ecma262/#sec-buffer-objects)  
> **Type Definition Source**: [microsoft/TypeScript lib.es2024.d.ts](https://github.com/microsoft/TypeScript/tree/main/src/lib)  
> **Gate Oracle**: TC39 Test262 Test Suite & TypeScript baselines

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
| `Buffer.alloc(size: number, fill?: number \| string): Buffer` | `alloc(size: number, fill?: number \| string): Buffer` | `__buffer.alloc` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.allocUnsafe(size: number): Buffer` | `allocUnsafe(size: number): Buffer` | `__buffer.allocUnsafe` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.compare(other: Buffer \| Uint8Array): number` | `compare(other: Buffer \| Uint8Array): number` | `__buffer.compare` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.concat(list: (Buffer \| Uint8Array)[], totalLength?: number): Buffer` | `concat(list: (Buffer \| Uint8Array)[], totalLength?: number): Buffer` | `__buffer.concat` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.copy(target: Buffer \| Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number` | `copy(target: Buffer \| Uint8Array, targetStart?: number, sourceStart?: number, sourceEnd?: number): number` | `__buffer.copy` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.equals(other: Buffer \| Uint8Array): boolean` | `equals(other: Buffer \| Uint8Array): boolean` | `__buffer.equals` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__buffer.fill` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.from(str: string, encoding?: string): Buffer` | `from(str: string, encoding?: string): Buffer` | `__buffer.from` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.indexOf(value: string \| number \| Uint8Array, byteOffset?: number, encoding?: string): number` | `indexOf(value: string \| number \| Uint8Array, byteOffset?: number, encoding?: string): number` | `__buffer.indexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.isBuffer(obj: any): boolean` | `isBuffer(obj: any): boolean` | `__buffer.isBuffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readDoubleBE(offset: number): number` | `readDoubleBE(offset: number): number` | `__buffer.readDoubleBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readDoubleLE(offset: number): number` | `readDoubleLE(offset: number): number` | `__buffer.readDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readFloatBE(offset: number): number` | `readFloatBE(offset: number): number` | `__buffer.readFloatBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readFloatLE(offset: number): number` | `readFloatLE(offset: number): number` | `__buffer.readFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readInt16BE(offset: number): number` | `readInt16BE(offset: number): number` | `__buffer.readInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readInt16LE(offset: number): number` | `readInt16LE(offset: number): number` | `__buffer.readInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readInt32BE(offset: number): number` | `readInt32BE(offset: number): number` | `__buffer.readInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readInt32LE(offset: number): number` | `readInt32LE(offset: number): number` | `__buffer.readInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readInt8(offset: number): number` | `readInt8(offset: number): number` | `__buffer.readInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readUInt16BE(offset: number): number` | `readUInt16BE(offset: number): number` | `__buffer.readUInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readUInt16LE(offset: number): number` | `readUInt16LE(offset: number): number` | `__buffer.readUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readUInt32BE(offset: number): number` | `readUInt32BE(offset: number): number` | `__buffer.readUInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readUInt32LE(offset: number): number` | `readUInt32LE(offset: number): number` | `__buffer.readUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readUInt8(offset: number): number` | `readUInt8(offset: number): number` | `__buffer.readUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readonly buffer: ArrayBuffer` | `readonly buffer: ArrayBuffer` | `__buffer.buffer` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readonly byteLength: number` | `readonly byteLength: number` | `__buffer.byteLength` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readonly byteOffset: number` | `readonly byteOffset: number` | `__buffer.byteOffset` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.readonly length: number` | `readonly length: number` | `__buffer.length` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.set(array: ArrayLike<number> \| Array<number> \| Uint8Array, offset?: number): void` | `set(array: ArrayLike<number> \| Array<number> \| Uint8Array, offset?: number): void` | `__buffer.set` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.slice(begin?: number, end?: number): Buffer` | `slice(begin?: number, end?: number): Buffer` | `__buffer.slice` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.subarray(begin?: number, end?: number): Buffer` | `subarray(begin?: number, end?: number): Buffer` | `__buffer.subarray` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.toString(encoding?: string, start?: number, end?: number): string` | `toString(encoding?: string, start?: number, end?: number): string` | `__buffer.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeDoubleBE(value: number, offset: number): number` | `writeDoubleBE(value: number, offset: number): number` | `__buffer.writeDoubleBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeDoubleLE(value: number, offset: number): number` | `writeDoubleLE(value: number, offset: number): number` | `__buffer.writeDoubleLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeFloatBE(value: number, offset: number): number` | `writeFloatBE(value: number, offset: number): number` | `__buffer.writeFloatBE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeFloatLE(value: number, offset: number): number` | `writeFloatLE(value: number, offset: number): number` | `__buffer.writeFloatLE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeInt16BE(value: number, offset: number): number` | `writeInt16BE(value: number, offset: number): number` | `__buffer.writeInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeInt16LE(value: number, offset: number): number` | `writeInt16LE(value: number, offset: number): number` | `__buffer.writeInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeInt32BE(value: number, offset: number): number` | `writeInt32BE(value: number, offset: number): number` | `__buffer.writeInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeInt32LE(value: number, offset: number): number` | `writeInt32LE(value: number, offset: number): number` | `__buffer.writeInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeInt8(value: number, offset: number): number` | `writeInt8(value: number, offset: number): number` | `__buffer.writeInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeUInt16BE(value: number, offset: number): number` | `writeUInt16BE(value: number, offset: number): number` | `__buffer.writeUInt16BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeUInt16LE(value: number, offset: number): number` | `writeUInt16LE(value: number, offset: number): number` | `__buffer.writeUInt16LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeUInt32BE(value: number, offset: number): number` | `writeUInt32BE(value: number, offset: number): number` | `__buffer.writeUInt32BE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeUInt32LE(value: number, offset: number): number` | `writeUInt32LE(value: number, offset: number): number` | `__buffer.writeUInt32LE` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |
| `Buffer.writeUInt8(value: number, offset: number): number` | `writeUInt8(value: number, offset: number): number` | `__buffer.writeUInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/buffer.ts` |

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
