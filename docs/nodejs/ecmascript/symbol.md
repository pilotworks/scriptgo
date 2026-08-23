# Symbol Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Symbol Specification](https://tc39.es/ecma262/#sec-symbol-objects)  
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
| `Symbol.for(key: string): symbol` | `for(key: string): symbol` | `__symbol.for` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.keyFor(sym: symbol): string \| undefined` | `keyFor(sym: symbol): string \| undefined` | `__symbol.keyFor` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly asyncDispose: unique symbol` | `readonly asyncDispose: unique symbol` | `__symbol.asyncDispose` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly asyncIterator: unique symbol` | `readonly asyncIterator: unique symbol` | `__symbol.asyncIterator` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly description: string \| undefined` | `readonly description: string \| undefined` | `__symbol.description` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly dispose: unique symbol` | `readonly dispose: unique symbol` | `__symbol.dispose` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly hasInstance: unique symbol` | `readonly hasInstance: unique symbol` | `__symbol.hasInstance` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly isConcatSpreadable: unique symbol` | `readonly isConcatSpreadable: unique symbol` | `__symbol.isConcatSpreadable` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly iterator: unique symbol` | `readonly iterator: unique symbol` | `__symbol.iterator` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly match: unique symbol` | `readonly match: unique symbol` | `__symbol.match` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly matchAll: unique symbol` | `readonly matchAll: unique symbol` | `__symbol.matchAll` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly metadata: unique symbol` | `readonly metadata: unique symbol` | `__symbol.metadata` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly replace: unique symbol` | `readonly replace: unique symbol` | `__symbol.replace` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly search: unique symbol` | `readonly search: unique symbol` | `__symbol.search` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly species: unique symbol` | `readonly species: unique symbol` | `__symbol.species` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly split: unique symbol` | `readonly split: unique symbol` | `__symbol.split` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly toPrimitive: unique symbol` | `readonly toPrimitive: unique symbol` | `__symbol.toPrimitive` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly toStringTag: unique symbol` | `readonly toStringTag: unique symbol` | `__symbol.toStringTag` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.readonly unscopables: unique symbol` | `readonly unscopables: unique symbol` | `__symbol.unscopables` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.toString(): string` | `toString(): string` | `__symbol.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |
| `Symbol.valueOf(): symbol` | `valueOf(): symbol` | `__symbol.valueOf` | ✅ Done | `internal/compiler/testdata/corpus/api/symbol.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `symbol` are organized per API under `internal/compiler/testdata/corpus/symbol/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/symbol/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
