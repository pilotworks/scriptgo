# Set Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Set Specification](https://tc39.es/ecma262/#sec-set-objects)  
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
| `Set.add(value: T): this` | `add(value: T): this` | `__set.add` | ✅ Done | `internal/compiler/testdata/corpus/api/set.ts` |
| `Set.clear(): void` | `clear(): void` | `__set.clear` | ✅ Done | `internal/compiler/testdata/corpus/api/set.ts` |
| `Set.delete(value: T): boolean` | `delete(value: T): boolean` | `__set.delete` | ✅ Done | `internal/compiler/testdata/corpus/api/set.ts` |
| `Set.forEach(callbackfn: (value: T, value2: T, set: Set<T>) => void, thisArg?: any): void` | `forEach(callbackfn: (value: T, value2: T, set: Set<T>) => void, thisArg?: any): void` | `__set.forEach` | ✅ Done | `internal/compiler/testdata/corpus/api/set.ts` |
| `Set.has(value: T): boolean` | `has(value: T): boolean` | `__set.has` | ✅ Done | `internal/compiler/testdata/corpus/api/set.ts` |
| `Set.difference<U>(other: ReadonlySetLike<U>): Set<T>` | `difference<U>(other: ReadonlySetLike<U>): Set<T>` | `__set.difference` | 📋 Planned | - |
| `Set.entries(): SetIterator<[T, T]>` | `entries(): SetIterator<[T, T]>` | `__set.entries` | 📋 Planned | - |
| `Set.intersection<U>(other: ReadonlySetLike<U>): Set<T & U>` | `intersection<U>(other: ReadonlySetLike<U>): Set<T & U>` | `__set.intersection` | 📋 Planned | - |
| `Set.isDisjointFrom(other: ReadonlySetLike<unknown>): boolean` | `isDisjointFrom(other: ReadonlySetLike<unknown>): boolean` | `__set.isDisjointFrom` | 📋 Planned | - |
| `Set.isSubsetOf(other: ReadonlySetLike<unknown>): boolean` | `isSubsetOf(other: ReadonlySetLike<unknown>): boolean` | `__set.isSubsetOf` | 📋 Planned | - |
| `Set.isSupersetOf(other: ReadonlySetLike<unknown>): boolean` | `isSupersetOf(other: ReadonlySetLike<unknown>): boolean` | `__set.isSupersetOf` | 📋 Planned | - |
| `Set.keys(): SetIterator<T>` | `keys(): SetIterator<T>` | `__set.keys` | 📋 Planned | - |
| `Set.readonly size: number` | `readonly size: number` | `__set.size` | 📋 Planned | - |
| `Set.symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T \| U>` | `symmetricDifference<U>(other: ReadonlySetLike<U>): Set<T \| U>` | `__set.symmetricDifference` | 📋 Planned | - |
| `Set.union<U>(other: ReadonlySetLike<U>): Set<T \| U>` | `union<U>(other: ReadonlySetLike<U>): Set<T \| U>` | `__set.union` | 📋 Planned | - |
| `Set.values(): SetIterator<T>` | `values(): SetIterator<T>` | `__set.values` | 📋 Planned | - |
| `new Set<T = any>(values?: readonly T[] \| null): Set<T>` | `new <T = any>(values?: readonly T[] \| null): Set<T>` | `__set.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `set` are organized per API under `internal/compiler/testdata/corpus/set/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/set/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
