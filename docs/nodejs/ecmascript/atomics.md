# Atomics Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Atomics Specification](https://tc39.es/ecma262/#sec-atomics-objects)  
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
| `Atomics.add(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `add(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.add` | 📋 Planned | - |
| `Atomics.and(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `and(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.and` | 📋 Planned | - |
| `Atomics.compareExchange(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, expectedValue: number, replacementValue: number): number` | `compareExchange(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, expectedValue: number, replacementValue: number): number` | `__atomics.compareExchange` | 📋 Planned | - |
| `Atomics.exchange(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `exchange(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.exchange` | 📋 Planned | - |
| `Atomics.isLockFree(size: number): boolean` | `isLockFree(size: number): boolean` | `__atomics.isLockFree` | 📋 Planned | - |
| `Atomics.load(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number): number` | `load(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number): number` | `__atomics.load` | 📋 Planned | - |
| `Atomics.notify(typedArray: Int32Array<ArrayBufferLike>, index: number, count?: number): number` | `notify(typedArray: Int32Array<ArrayBufferLike>, index: number, count?: number): number` | `__atomics.notify` | 📋 Planned | - |
| `Atomics.or(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `or(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.or` | 📋 Planned | - |
| `Atomics.pause(n?: number): void` | `pause(n?: number): void` | `__atomics.pause` | 📋 Planned | - |
| `Atomics.store(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `store(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.store` | 📋 Planned | - |
| `Atomics.sub(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `sub(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.sub` | 📋 Planned | - |
| `Atomics.wait(typedArray: Int32Array<ArrayBufferLike>, index: number, value: number, timeout?: number): "ok" \| "not-equal" \| "timed-out"` | `wait(typedArray: Int32Array<ArrayBufferLike>, index: number, value: number, timeout?: number): "ok" \| "not-equal" \| "timed-out"` | `__atomics.wait` | 📋 Planned | - |
| `Atomics.waitAsync(typedArray: Int32Array, index: number, value: number, timeout?: number): { async: false; value: "not-equal" \| "timed-out"; } \| { async: true; value: Promise<"ok" \| "timed-out">; }` | `waitAsync(typedArray: Int32Array, index: number, value: number, timeout?: number): { async: false; value: "not-equal" \| "timed-out"; } \| { async: true; value: Promise<"ok" \| "timed-out">; }` | `__atomics.waitAsync` | 📋 Planned | - |
| `Atomics.xor(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `xor(typedArray: Int8Array<ArrayBufferLike> \| Uint8Array<ArrayBufferLike> \| Int16Array<ArrayBufferLike> \| Uint16Array<ArrayBufferLike> \| Int32Array<ArrayBufferLike> \| Uint32Array<ArrayBufferLike>, index: number, value: number): number` | `__atomics.xor` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `atomics` are organized per API under `internal/compiler/testdata/corpus/atomics/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/atomics/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
