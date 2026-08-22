# BigIntToLocaleStringOptions Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 BigIntToLocaleStringOptions Specification](https://tc39.es/ecma262/#sec-biginttolocalestringoptions-objects)  
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
| `BigIntToLocaleStringOptions.compactDisplay?: string` | `compactDisplay?: string` | `__biginttolocalestringoptions.compactDisplay` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.currency?: string` | `currency?: string` | `__biginttolocalestringoptions.currency` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.currencyDisplay?: string` | `currencyDisplay?: string` | `__biginttolocalestringoptions.currencyDisplay` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.localeMatcher?: string` | `localeMatcher?: string` | `__biginttolocalestringoptions.localeMatcher` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.maximumFractionDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20` | `maximumFractionDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20` | `__biginttolocalestringoptions.maximumFractionDigits` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.maximumSignificantDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `maximumSignificantDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `__biginttolocalestringoptions.maximumSignificantDigits` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.minimumFractionDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20` | `minimumFractionDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20` | `__biginttolocalestringoptions.minimumFractionDigits` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.minimumIntegerDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `minimumIntegerDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `__biginttolocalestringoptions.minimumIntegerDigits` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.minimumSignificantDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `minimumSignificantDigits?: 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| 10 \| 11 \| 12 \| 13 \| 14 \| 15 \| 16 \| 17 \| 18 \| 19 \| 20 \| 21` | `__biginttolocalestringoptions.minimumSignificantDigits` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.notation?: string` | `notation?: string` | `__biginttolocalestringoptions.notation` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.numberingSystem?: string` | `numberingSystem?: string` | `__biginttolocalestringoptions.numberingSystem` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.style?: string` | `style?: string` | `__biginttolocalestringoptions.style` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.unit?: string` | `unit?: string` | `__biginttolocalestringoptions.unit` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.unitDisplay?: string` | `unitDisplay?: string` | `__biginttolocalestringoptions.unitDisplay` | 📋 Planned | - |
| `BigIntToLocaleStringOptions.useGrouping?: boolean` | `useGrouping?: boolean` | `__biginttolocalestringoptions.useGrouping` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `biginttolocalestringoptions` are organized per API under `internal/compiler/testdata/corpus/biginttolocalestringoptions/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/biginttolocalestringoptions/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
