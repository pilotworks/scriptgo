# RegExp Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 RegExp Specification](https://tc39.es/ecma262/#sec-regexp-objects)  
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
| `RegExp.compile(pattern: string, flags?: string): this` | `compile(pattern: string, flags?: string): this` | `__regexp.compile` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.escape(string: string): string` | `escape(string: string): string` | `__regexp.escape` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.exec(string: string): RegExpExecArray \| null` | `exec(string: string): RegExpExecArray \| null` | `__regexp.exec` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.lastIndex: number` | `lastIndex: number` | `__regexp.lastIndex` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly dotAll: boolean` | `readonly dotAll: boolean` | `__regexp.dotAll` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly flags: string` | `readonly flags: string` | `__regexp.flags` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly global: boolean` | `readonly global: boolean` | `__regexp.global` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly hasIndices: boolean` | `readonly hasIndices: boolean` | `__regexp.hasIndices` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly ignoreCase: boolean` | `readonly ignoreCase: boolean` | `__regexp.ignoreCase` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly multiline: boolean` | `readonly multiline: boolean` | `__regexp.multiline` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly source: string` | `readonly source: string` | `__regexp.source` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly sticky: boolean` | `readonly sticky: boolean` | `__regexp.sticky` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly unicode: boolean` | `readonly unicode: boolean` | `__regexp.unicode` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.readonly unicodeSets: boolean` | `readonly unicodeSets: boolean` | `__regexp.unicodeSets` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `RegExp.test(string: string): boolean` | `test(string: string): boolean` | `__regexp.test` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |
| `new RegExp(pattern: RegExp \| string, flags?: string): RegExp` | `new (pattern: RegExp \| string, flags?: string): RegExp` | `__regexp.new` | ✅ Done | `internal/compiler/testdata/corpus/api/regexp.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `regexp` are organized per API under `internal/compiler/testdata/corpus/regexp/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/regexp/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
