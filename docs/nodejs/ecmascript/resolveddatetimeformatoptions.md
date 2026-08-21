# ResolvedDateTimeFormatOptions Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 ResolvedDateTimeFormatOptions Specification](https://tc39.es/ecma262/#sec-resolveddatetimeformatoptions-objects)  
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
| `ResolvedDateTimeFormatOptions.calendar: string` | `calendar: string` | `__resolveddatetimeformatoptions.calendar` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.dateStyle?: "full" \| "long" \| "medium" \| "short"` | `dateStyle?: "full" \| "long" \| "medium" \| "short"` | `__resolveddatetimeformatoptions.dateStyle?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.day?: string` | `day?: string` | `__resolveddatetimeformatoptions.day?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.dayPeriod?: "narrow" \| "short" \| "long"` | `dayPeriod?: "narrow" \| "short" \| "long"` | `__resolveddatetimeformatoptions.dayPeriod?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.era?: string` | `era?: string` | `__resolveddatetimeformatoptions.era?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.formatMatcher?: "basic" \| "best fit" \| "best fit"` | `formatMatcher?: "basic" \| "best fit" \| "best fit"` | `__resolveddatetimeformatoptions.formatMatcher?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.fractionalSecondDigits?: 1 \| 2 \| 3` | `fractionalSecondDigits?: 1 \| 2 \| 3` | `__resolveddatetimeformatoptions.fractionalSecondDigits?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.hour12?: boolean` | `hour12?: boolean` | `__resolveddatetimeformatoptions.hour12?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.hour?: string` | `hour?: string` | `__resolveddatetimeformatoptions.hour?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.hourCycle?: "h11" \| "h12" \| "h23" \| "h24"` | `hourCycle?: "h11" \| "h12" \| "h23" \| "h24"` | `__resolveddatetimeformatoptions.hourCycle?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.locale: string` | `locale: string` | `__resolveddatetimeformatoptions.locale` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.minute?: string` | `minute?: string` | `__resolveddatetimeformatoptions.minute?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.month?: string` | `month?: string` | `__resolveddatetimeformatoptions.month?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.numberingSystem: string` | `numberingSystem: string` | `__resolveddatetimeformatoptions.numberingSystem` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.second?: string` | `second?: string` | `__resolveddatetimeformatoptions.second?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.timeStyle?: "full" \| "long" \| "medium" \| "short"` | `timeStyle?: "full" \| "long" \| "medium" \| "short"` | `__resolveddatetimeformatoptions.timeStyle?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.timeZone: string` | `timeZone: string` | `__resolveddatetimeformatoptions.timeZone` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.timeZoneName?: string` | `timeZoneName?: string` | `__resolveddatetimeformatoptions.timeZoneName?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.weekday?: string` | `weekday?: string` | `__resolveddatetimeformatoptions.weekday?` | 📋 Planned | - |
| `ResolvedDateTimeFormatOptions.year?: string` | `year?: string` | `__resolveddatetimeformatoptions.year?` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `resolveddatetimeformatoptions` are organized per API under `internal/compiler/testdata/corpus/resolveddatetimeformatoptions/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/resolveddatetimeformatoptions/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
