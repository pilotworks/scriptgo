# Promise Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Promise Specification](https://tc39.es/ecma262/#sec-promise-objects)  
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
| `Promise.all<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>[]>` | `all<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>[]>` | `__promise.all<T>` | 📋 Planned | - |
| `Promise.allSettled<T>(values: Iterable<T \| PromiseLike<T>>): Promise<PromiseSettledResult<Awaited<T>>[]>` | `allSettled<T>(values: Iterable<T \| PromiseLike<T>>): Promise<PromiseSettledResult<Awaited<T>>[]>` | `__promise.allSettled<T>` | 📋 Planned | - |
| `Promise.any<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>>` | `any<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>>` | `__promise.any<T>` | 📋 Planned | - |
| `Promise.finally(onfinally?: (() => void) \| undefined \| null): Promise<T>` | `finally(onfinally?: (() => void) \| undefined \| null): Promise<T>` | `__promise.finally` | 📋 Planned | - |
| `Promise.race<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>>` | `race<T>(values: Iterable<T \| PromiseLike<T>>): Promise<Awaited<T>>` | `__promise.race<T>` | 📋 Planned | - |
| `Promise.readonly prototype: Promise<any>` | `readonly prototype: Promise<any>` | `__promise.prototype` | 📋 Planned | - |
| `Promise.resolve(): Promise<void>` | `resolve(): Promise<void>` | `__promise.resolve` | 📋 Planned | - |
| `Promise.resolve<T>(value: T): Promise<Awaited<T>>` | `resolve<T>(value: T): Promise<Awaited<T>>` | `__promise.resolve<T>` | 📋 Planned | - |
| `Promise.withResolvers<T>(): PromiseWithResolvers<T>` | `withResolvers<T>(): PromiseWithResolvers<T>` | `__promise.withResolvers<T>` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `promise` are organized per API under `internal/compiler/testdata/corpus/promise/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/promise/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
