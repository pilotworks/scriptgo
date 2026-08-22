# NumberFormatOptions Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 NumberFormatOptions Specification](https://tc39.es/ecma262/#sec-numberformatoptions-objects)  
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
| `NumberFormatOptions.compactDisplay?: "short" \| "long" \| undefined` | `compactDisplay?: "short" \| "long" \| undefined` | `__numberformatoptions.compactDisplay` | 📋 Planned | - |
| `NumberFormatOptions.currency?: string \| undefined` | `currency?: string \| undefined` | `__numberformatoptions.currency` | 📋 Planned | - |
| `NumberFormatOptions.currencyDisplay?: NumberFormatOptionsCurrencyDisplay \| undefined` | `currencyDisplay?: NumberFormatOptionsCurrencyDisplay \| undefined` | `__numberformatoptions.currencyDisplay` | 📋 Planned | - |
| `NumberFormatOptions.currencySign?: "standard" \| "accounting" \| undefined` | `currencySign?: "standard" \| "accounting" \| undefined` | `__numberformatoptions.currencySign` | 📋 Planned | - |
| `NumberFormatOptions.localeMatcher?: "lookup" \| "best fit" \| undefined` | `localeMatcher?: "lookup" \| "best fit" \| undefined` | `__numberformatoptions.localeMatcher` | 📋 Planned | - |
| `NumberFormatOptions.maximumFractionDigits?: number \| undefined` | `maximumFractionDigits?: number \| undefined` | `__numberformatoptions.maximumFractionDigits` | 📋 Planned | - |
| `NumberFormatOptions.maximumSignificantDigits?: number \| undefined` | `maximumSignificantDigits?: number \| undefined` | `__numberformatoptions.maximumSignificantDigits` | 📋 Planned | - |
| `NumberFormatOptions.minimumFractionDigits?: number \| undefined` | `minimumFractionDigits?: number \| undefined` | `__numberformatoptions.minimumFractionDigits` | 📋 Planned | - |
| `NumberFormatOptions.minimumIntegerDigits?: number \| undefined` | `minimumIntegerDigits?: number \| undefined` | `__numberformatoptions.minimumIntegerDigits` | 📋 Planned | - |
| `NumberFormatOptions.minimumSignificantDigits?: number \| undefined` | `minimumSignificantDigits?: number \| undefined` | `__numberformatoptions.minimumSignificantDigits` | 📋 Planned | - |
| `NumberFormatOptions.notation?: "standard" \| "scientific" \| "engineering" \| "compact" \| undefined` | `notation?: "standard" \| "scientific" \| "engineering" \| "compact" \| undefined` | `__numberformatoptions.notation` | 📋 Planned | - |
| `NumberFormatOptions.numberingSystem?: string \| undefined` | `numberingSystem?: string \| undefined` | `__numberformatoptions.numberingSystem` | 📋 Planned | - |
| `NumberFormatOptions.roundingIncrement?: 1 \| 2 \| 5 \| 10 \| 20 \| 25 \| 50 \| 100 \| 200 \| 250 \| 500 \| 1000 \| 2000 \| 2500 \| 5000 \| undefined` | `roundingIncrement?: 1 \| 2 \| 5 \| 10 \| 20 \| 25 \| 50 \| 100 \| 200 \| 250 \| 500 \| 1000 \| 2000 \| 2500 \| 5000 \| undefined` | `__numberformatoptions.roundingIncrement` | 📋 Planned | - |
| `NumberFormatOptions.roundingMode?: "ceil" \| "floor" \| "expand" \| "trunc" \| "halfCeil" \| "halfFloor" \| "halfExpand" \| "halfTrunc" \| "halfEven" \| undefined` | `roundingMode?: "ceil" \| "floor" \| "expand" \| "trunc" \| "halfCeil" \| "halfFloor" \| "halfExpand" \| "halfTrunc" \| "halfEven" \| undefined` | `__numberformatoptions.roundingMode` | 📋 Planned | - |
| `NumberFormatOptions.roundingPriority?: "auto" \| "morePrecision" \| "lessPrecision" \| undefined` | `roundingPriority?: "auto" \| "morePrecision" \| "lessPrecision" \| undefined` | `__numberformatoptions.roundingPriority` | 📋 Planned | - |
| `NumberFormatOptions.signDisplay?: NumberFormatOptionsSignDisplay \| undefined` | `signDisplay?: NumberFormatOptionsSignDisplay \| undefined` | `__numberformatoptions.signDisplay` | 📋 Planned | - |
| `NumberFormatOptions.style?: NumberFormatOptionsStyle \| undefined` | `style?: NumberFormatOptionsStyle \| undefined` | `__numberformatoptions.style` | 📋 Planned | - |
| `NumberFormatOptions.trailingZeroDisplay?: "auto" \| "stripIfInteger" \| undefined` | `trailingZeroDisplay?: "auto" \| "stripIfInteger" \| undefined` | `__numberformatoptions.trailingZeroDisplay` | 📋 Planned | - |
| `NumberFormatOptions.unit?: string \| undefined` | `unit?: string \| undefined` | `__numberformatoptions.unit` | 📋 Planned | - |
| `NumberFormatOptions.unitDisplay?: "short" \| "long" \| "narrow" \| undefined` | `unitDisplay?: "short" \| "long" \| "narrow" \| undefined` | `__numberformatoptions.unitDisplay` | 📋 Planned | - |
| `NumberFormatOptions.useGrouping?: NumberFormatOptionsUseGrouping \| undefined` | `useGrouping?: NumberFormatOptionsUseGrouping \| undefined` | `__numberformatoptions.useGrouping` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `numberformatoptions` are organized per API under `internal/compiler/testdata/corpus/numberformatoptions/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/numberformatoptions/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
