# DurationFormatOptions Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 DurationFormatOptions Specification](https://tc39.es/ecma262/#sec-durationformatoptions-objects)  
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
| `DurationFormatOptions.days?: "long" \| "short" \| "narrow" \| undefined` | `days?: "long" \| "short" \| "narrow" \| undefined` | `__durationformatoptions.days?` | 📋 Planned | - |
| `DurationFormatOptions.daysDisplay?: DurationFormatDisplayOption \| undefined` | `daysDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.daysDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.fractionalDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| undefined` | `fractionalDigits?: 0 \| 1 \| 2 \| 3 \| 4 \| 5 \| 6 \| 7 \| 8 \| 9 \| undefined` | `__durationformatoptions.fractionalDigits?` | 📋 Planned | - |
| `DurationFormatOptions.hours?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `hours?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `__durationformatoptions.hours?` | 📋 Planned | - |
| `DurationFormatOptions.hoursDisplay?: DurationFormatDisplayOption \| undefined` | `hoursDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.hoursDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.localeMatcher?: DurationFormatLocaleMatcher \| undefined` | `localeMatcher?: DurationFormatLocaleMatcher \| undefined` | `__durationformatoptions.localeMatcher?` | 📋 Planned | - |
| `DurationFormatOptions.microseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `microseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `__durationformatoptions.microseconds?` | 📋 Planned | - |
| `DurationFormatOptions.microsecondsDisplay?: DurationFormatDisplayOption \| undefined` | `microsecondsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.microsecondsDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.milliseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `milliseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `__durationformatoptions.milliseconds?` | 📋 Planned | - |
| `DurationFormatOptions.millisecondsDisplay?: DurationFormatDisplayOption \| undefined` | `millisecondsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.millisecondsDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.minutes?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `minutes?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `__durationformatoptions.minutes?` | 📋 Planned | - |
| `DurationFormatOptions.minutesDisplay?: DurationFormatDisplayOption \| undefined` | `minutesDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.minutesDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.months?: "long" \| "short" \| "narrow" \| undefined` | `months?: "long" \| "short" \| "narrow" \| undefined` | `__durationformatoptions.months?` | 📋 Planned | - |
| `DurationFormatOptions.monthsDisplay?: DurationFormatDisplayOption \| undefined` | `monthsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.monthsDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.nanoseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `nanoseconds?: "long" \| "short" \| "narrow" \| "numeric" \| undefined` | `__durationformatoptions.nanoseconds?` | 📋 Planned | - |
| `DurationFormatOptions.nanosecondsDisplay?: DurationFormatDisplayOption \| undefined` | `nanosecondsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.nanosecondsDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.numberingSystem?: string \| undefined` | `numberingSystem?: string \| undefined` | `__durationformatoptions.numberingSystem?` | 📋 Planned | - |
| `DurationFormatOptions.seconds?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `seconds?: "long" \| "short" \| "narrow" \| "numeric" \| "2-digit" \| undefined` | `__durationformatoptions.seconds?` | 📋 Planned | - |
| `DurationFormatOptions.secondsDisplay?: DurationFormatDisplayOption \| undefined` | `secondsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.secondsDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.style?: DurationFormatStyle \| undefined` | `style?: DurationFormatStyle \| undefined` | `__durationformatoptions.style?` | 📋 Planned | - |
| `DurationFormatOptions.weeks?: "long" \| "short" \| "narrow" \| undefined` | `weeks?: "long" \| "short" \| "narrow" \| undefined` | `__durationformatoptions.weeks?` | 📋 Planned | - |
| `DurationFormatOptions.weeksDisplay?: DurationFormatDisplayOption \| undefined` | `weeksDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.weeksDisplay?` | 📋 Planned | - |
| `DurationFormatOptions.years?: "long" \| "short" \| "narrow" \| undefined` | `years?: "long" \| "short" \| "narrow" \| undefined` | `__durationformatoptions.years?` | 📋 Planned | - |
| `DurationFormatOptions.yearsDisplay?: DurationFormatDisplayOption \| undefined` | `yearsDisplay?: DurationFormatDisplayOption \| undefined` | `__durationformatoptions.yearsDisplay?` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `durationformatoptions` are organized per API under `internal/compiler/testdata/corpus/durationformatoptions/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/durationformatoptions/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
