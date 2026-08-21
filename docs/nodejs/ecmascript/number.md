# Number Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Number Specification](https://tc39.es/ecma262/#sec-number-objects)  
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
| `Number.isFinite(number: unknown): boolean` | `isFinite(number: unknown): boolean` | `__number.isFinite` | ✅ Done | `internal/compiler/testdata/corpus/api/number/isFinite/` |
| `Number.isInteger(number: unknown): boolean` | `isInteger(number: unknown): boolean` | `__number.isInteger` | ✅ Done | `internal/compiler/testdata/corpus/api/number/isInteger/` |
| `Number.isNaN(number: unknown): boolean` | `isNaN(number: unknown): boolean` | `__number.isNaN` | ✅ Done | `internal/compiler/testdata/corpus/api/number/isNaN/` |
| `Number.parseFloat(string: string): number` | `parseFloat(string: string): number` | `__number.parseFloat` | ✅ Done | `internal/compiler/testdata/corpus/api/number/parseFloat/` |
| `Number.parseInt(string: string, radix?: number): number` | `parseInt(string: string, radix?: number): number` | `__number.parseInt` | ✅ Done | `internal/compiler/testdata/corpus/api/number/parseInt/` |
| `Number.isSafeInteger(number: unknown): boolean` | `isSafeInteger(number: unknown): boolean` | `__number.isSafeInteger` | 📋 Planned | - |
| `Number.new (value?: any): Number` | `new (value?: any): Number` | `__number.new` | 📋 Planned | - |
| `Number.readonly EPSILON: number` | `readonly EPSILON: number` | `__number.EPSILON` | 📋 Planned | - |
| `Number.readonly MAX_SAFE_INTEGER: number` | `readonly MAX_SAFE_INTEGER: number` | `__number.MAX_SAFE_INTEGER` | 📋 Planned | - |
| `Number.readonly MAX_VALUE: number` | `readonly MAX_VALUE: number` | `__number.MAX_VALUE` | 📋 Planned | - |
| `Number.readonly MIN_SAFE_INTEGER: number` | `readonly MIN_SAFE_INTEGER: number` | `__number.MIN_SAFE_INTEGER` | 📋 Planned | - |
| `Number.readonly MIN_VALUE: number` | `readonly MIN_VALUE: number` | `__number.MIN_VALUE` | 📋 Planned | - |
| `Number.readonly NEGATIVE_INFINITY: number` | `readonly NEGATIVE_INFINITY: number` | `__number.NEGATIVE_INFINITY` | 📋 Planned | - |
| `Number.readonly NaN: number` | `readonly NaN: number` | `__number.NaN` | 📋 Planned | - |
| `Number.readonly POSITIVE_INFINITY: number` | `readonly POSITIVE_INFINITY: number` | `__number.POSITIVE_INFINITY` | 📋 Planned | - |
| `Number.readonly prototype: Number` | `readonly prototype: Number` | `__number.prototype` | 📋 Planned | - |
| `Number.toExponential(fractionDigits?: number): string` | `toExponential(fractionDigits?: number): string` | `__number.toExponential` | 📋 Planned | - |
| `Number.toFixed(fractionDigits?: number): string` | `toFixed(fractionDigits?: number): string` | `__number.toFixed` | 📋 Planned | - |
| `Number.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.NumberFormatOptions): string` | `__number.toLocaleString` | 📋 Planned | - |
| `Number.toPrecision(precision?: number): string` | `toPrecision(precision?: number): string` | `__number.toPrecision` | 📋 Planned | - |
| `Number.toString(radix?: number): string` | `toString(radix?: number): string` | `__number.toString` | 📋 Planned | - |
| `Number.valueOf(): number` | `valueOf(): number` | `__number.valueOf` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `number` are organized per API under `internal/compiler/testdata/corpus/number/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/number/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
