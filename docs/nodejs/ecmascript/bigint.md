# BigInt Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 BigInt Specification](https://tc39.es/ecma262/#sec-bigint-objects)  
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
| `BigInt.asIntN(bits: number, int: bigint): bigint` | `asIntN(bits: number, int: bigint): bigint` | `__bigint.asIntN` | ✅ Done | `internal/compiler/testdata/corpus/api/bigint.ts` |
| `BigInt.asUintN(bits: number, int: bigint): bigint` | `asUintN(bits: number, int: bigint): bigint` | `__bigint.asUintN` | ✅ Done | `internal/compiler/testdata/corpus/api/bigint.ts` |
| `BigInt.toLocaleString(locales?: Intl.LocalesArgument, options?: BigIntToLocaleStringOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: BigIntToLocaleStringOptions): string` | `__bigint.toLocaleString` | ✅ Done | `internal/compiler/testdata/corpus/api/bigint.ts` |
| `BigInt.toString(radix?: number): string` | `toString(radix?: number): string` | `__bigint.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/bigint.ts` |
| `BigInt.valueOf(): bigint` | `valueOf(): bigint` | `__bigint.valueOf` | ✅ Done | `internal/compiler/testdata/corpus/api/bigint.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `bigint` are organized per API under `internal/compiler/testdata/corpus/bigint/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/bigint/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
