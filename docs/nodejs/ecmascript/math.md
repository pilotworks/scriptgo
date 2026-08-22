# Math Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Math Specification](https://tc39.es/ecma262/#sec-math-objects)  
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
| `Math.abs(x: number): number` | `abs(x: number): number` | `__math.abs` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.atan(x: number): number` | `atan(x: number): number` | `__math.atan` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.atan2(y: number, x: number): number` | `atan2(y: number, x: number): number` | `__math.atan2` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.ceil(x: number): number` | `ceil(x: number): number` | `__math.ceil` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.cos(x: number): number` | `cos(x: number): number` | `__math.cos` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.exp(x: number): number` | `exp(x: number): number` | `__math.exp` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.floor(x: number): number` | `floor(x: number): number` | `__math.floor` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.hypot(...values: number[]): number` | `hypot(...values: number[]): number` | `__math.hypot` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.log(x: number): number` | `log(x: number): number` | `__math.log` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.log10(x: number): number` | `log10(x: number): number` | `__math.log10` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.log2(x: number): number` | `log2(x: number): number` | `__math.log2` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.max(...values: number[]): number` | `max(...values: number[]): number` | `__math.max` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.min(...values: number[]): number` | `min(...values: number[]): number` | `__math.min` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.pow(x: number, y: number): number` | `pow(x: number, y: number): number` | `__math.pow` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.random(): number` | `random(): number` | `__math.random` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.round(x: number): number` | `round(x: number): number` | `__math.round` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.sign(x: number): number` | `sign(x: number): number` | `__math.sign` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.sin(x: number): number` | `sin(x: number): number` | `__math.sin` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.sqrt(x: number): number` | `sqrt(x: number): number` | `__math.sqrt` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.tan(x: number): number` | `tan(x: number): number` | `__math.tan` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.trunc(x: number): number` | `trunc(x: number): number` | `__math.trunc` | ✅ Done | `internal/compiler/testdata/corpus/api/math.ts` |
| `Math.acos(x: number): number` | `acos(x: number): number` | `__math.acos` | 📋 Planned | - |
| `Math.acosh(x: number): number` | `acosh(x: number): number` | `__math.acosh` | 📋 Planned | - |
| `Math.asin(x: number): number` | `asin(x: number): number` | `__math.asin` | 📋 Planned | - |
| `Math.asinh(x: number): number` | `asinh(x: number): number` | `__math.asinh` | 📋 Planned | - |
| `Math.atanh(x: number): number` | `atanh(x: number): number` | `__math.atanh` | 📋 Planned | - |
| `Math.cbrt(x: number): number` | `cbrt(x: number): number` | `__math.cbrt` | 📋 Planned | - |
| `Math.clz32(x: number): number` | `clz32(x: number): number` | `__math.clz32` | 📋 Planned | - |
| `Math.cosh(x: number): number` | `cosh(x: number): number` | `__math.cosh` | 📋 Planned | - |
| `Math.expm1(x: number): number` | `expm1(x: number): number` | `__math.expm1` | 📋 Planned | - |
| `Math.f16round(x: number): number` | `f16round(x: number): number` | `__math.f16round` | 📋 Planned | - |
| `Math.fround(x: number): number` | `fround(x: number): number` | `__math.fround` | 📋 Planned | - |
| `Math.imul(x: number, y: number): number` | `imul(x: number, y: number): number` | `__math.imul` | 📋 Planned | - |
| `Math.log1p(x: number): number` | `log1p(x: number): number` | `__math.log1p` | 📋 Planned | - |
| `Math.readonly E: number` | `readonly E: number` | `__math.E` | 📋 Planned | - |
| `Math.readonly LN10: number` | `readonly LN10: number` | `__math.LN10` | 📋 Planned | - |
| `Math.readonly LN2: number` | `readonly LN2: number` | `__math.LN2` | 📋 Planned | - |
| `Math.readonly LOG10E: number` | `readonly LOG10E: number` | `__math.LOG10E` | 📋 Planned | - |
| `Math.readonly LOG2E: number` | `readonly LOG2E: number` | `__math.LOG2E` | 📋 Planned | - |
| `Math.readonly PI: number` | `readonly PI: number` | `__math.PI` | 📋 Planned | - |
| `Math.readonly SQRT1_2: number` | `readonly SQRT1_2: number` | `__math.SQRT1_2` | 📋 Planned | - |
| `Math.readonly SQRT2: number` | `readonly SQRT2: number` | `__math.SQRT2` | 📋 Planned | - |
| `Math.sinh(x: number): number` | `sinh(x: number): number` | `__math.sinh` | 📋 Planned | - |
| `Math.tanh(x: number): number` | `tanh(x: number): number` | `__math.tanh` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `math` are organized per API under `internal/compiler/testdata/corpus/math/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/math/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
