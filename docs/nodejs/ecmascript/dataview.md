# DataView Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 DataView Specification](https://tc39.es/ecma262/#sec-dataview-objects)  
> **Type Definition Source**: [microsoft/TypeScript lib.es2024.d.ts](https://github.com/microsoft/TypeScript/tree/main/src/lib)  
> **Gate Oracle**: TC39 Test262 Test Suite & TypeScript baselines

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
| `DataView.getFloat64(byteOffset: number, littleEndian?: boolean): number` | `getFloat64(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getFloat64` | ✅ Done | `internal/compiler/testdata/corpus/api/dataview.ts` |
| `DataView.getInt32(byteOffset: number, littleEndian?: boolean): number` | `getInt32(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getInt32` | ✅ Done | `internal/compiler/testdata/corpus/api/dataview.ts` |
| `DataView.getInt8(byteOffset: number): number` | `getInt8(byteOffset: number): number` | `__dataview.getInt8` | ✅ Done | `internal/compiler/testdata/corpus/api/dataview.ts` |
| `DataView.getBigInt64(byteOffset: number, littleEndian?: boolean): bigint` | `getBigInt64(byteOffset: number, littleEndian?: boolean): bigint` | `__dataview.getBigInt64` | 📋 Planned | - |
| `DataView.getBigUint64(byteOffset: number, littleEndian?: boolean): bigint` | `getBigUint64(byteOffset: number, littleEndian?: boolean): bigint` | `__dataview.getBigUint64` | 📋 Planned | - |
| `DataView.getFloat16(byteOffset: number, littleEndian?: boolean): number` | `getFloat16(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getFloat16` | 📋 Planned | - |
| `DataView.getFloat32(byteOffset: number, littleEndian?: boolean): number` | `getFloat32(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getFloat32` | 📋 Planned | - |
| `DataView.getInt16(byteOffset: number, littleEndian?: boolean): number` | `getInt16(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getInt16` | 📋 Planned | - |
| `DataView.getUint16(byteOffset: number, littleEndian?: boolean): number` | `getUint16(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getUint16` | 📋 Planned | - |
| `DataView.getUint32(byteOffset: number, littleEndian?: boolean): number` | `getUint32(byteOffset: number, littleEndian?: boolean): number` | `__dataview.getUint32` | 📋 Planned | - |
| `DataView.getUint8(byteOffset: number): number` | `getUint8(byteOffset: number): number` | `__dataview.getUint8` | 📋 Planned | - |
| `DataView.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__dataview.buffer` | 📋 Planned | - |
| `DataView.readonly byteLength: number` | `readonly byteLength: number` | `__dataview.byteLength` | 📋 Planned | - |
| `DataView.readonly byteOffset: number` | `readonly byteOffset: number` | `__dataview.byteOffset` | 📋 Planned | - |
| `DataView.setBigInt64(byteOffset: number, value: bigint, littleEndian?: boolean): void` | `setBigInt64(byteOffset: number, value: bigint, littleEndian?: boolean): void` | `__dataview.setBigInt64` | 📋 Planned | - |
| `DataView.setBigUint64(byteOffset: number, value: bigint, littleEndian?: boolean): void` | `setBigUint64(byteOffset: number, value: bigint, littleEndian?: boolean): void` | `__dataview.setBigUint64` | 📋 Planned | - |
| `DataView.setFloat16(byteOffset: number, value: number, littleEndian?: boolean): void` | `setFloat16(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setFloat16` | 📋 Planned | - |
| `DataView.setFloat32(byteOffset: number, value: number, littleEndian?: boolean): void` | `setFloat32(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setFloat32` | 📋 Planned | - |
| `DataView.setFloat64(byteOffset: number, value: number, littleEndian?: boolean): void` | `setFloat64(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setFloat64` | 📋 Planned | - |
| `DataView.setInt16(byteOffset: number, value: number, littleEndian?: boolean): void` | `setInt16(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setInt16` | 📋 Planned | - |
| `DataView.setInt32(byteOffset: number, value: number, littleEndian?: boolean): void` | `setInt32(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setInt32` | 📋 Planned | - |
| `DataView.setInt8(byteOffset: number, value: number): void` | `setInt8(byteOffset: number, value: number): void` | `__dataview.setInt8` | 📋 Planned | - |
| `DataView.setUint16(byteOffset: number, value: number, littleEndian?: boolean): void` | `setUint16(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setUint16` | 📋 Planned | - |
| `DataView.setUint32(byteOffset: number, value: number, littleEndian?: boolean): void` | `setUint32(byteOffset: number, value: number, littleEndian?: boolean): void` | `__dataview.setUint32` | 📋 Planned | - |
| `DataView.setUint8(byteOffset: number, value: number): void` | `setUint8(byteOffset: number, value: number): void` | `__dataview.setUint8` | 📋 Planned | - |
| `new DataView<TArrayBuffer extends ArrayBufferLike & { BYTES_PER_ELEMENT?: never; }>(buffer: TArrayBuffer, byteOffset?: number, byteLength?: number): DataView<TArrayBuffer>` | `new <TArrayBuffer extends ArrayBufferLike & { BYTES_PER_ELEMENT?: never; }>(buffer: TArrayBuffer, byteOffset?: number, byteLength?: number): DataView<TArrayBuffer>` | `__dataview.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `dataview` are organized per API under `internal/compiler/testdata/corpus/dataview/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/dataview/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
