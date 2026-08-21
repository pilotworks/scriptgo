# DateTimeFormatOptions Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 DateTimeFormatOptions Specification](https://tc39.es/ecma262/#sec-datetimeformatoptions-objects)  
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
| `DateTimeFormatOptions.calendar?: string \| undefined` | `calendar?: string \| undefined` | `__datetimeformatoptions.calendar?` | 📋 Planned | - |
| `DateTimeFormatOptions.dateStyle?: "full" \| "long" \| "medium" \| "short" \| undefined` | `dateStyle?: "full" \| "long" \| "medium" \| "short" \| undefined` | `__datetimeformatoptions.dateStyle?` | 📋 Planned | - |
| `DateTimeFormatOptions.day?: "numeric" \| "2-digit" \| undefined` | `day?: "numeric" \| "2-digit" \| undefined` | `__datetimeformatoptions.day?` | 📋 Planned | - |
| `DateTimeFormatOptions.dayPeriod?: "narrow" \| "short" \| "long" \| undefined` | `dayPeriod?: "narrow" \| "short" \| "long" \| undefined` | `__datetimeformatoptions.dayPeriod?` | 📋 Planned | - |
| `DateTimeFormatOptions.era?: "long" \| "short" \| "narrow" \| undefined` | `era?: "long" \| "short" \| "narrow" \| undefined` | `__datetimeformatoptions.era?` | 📋 Planned | - |
| `DateTimeFormatOptions.formatMatcher?: "basic" \| "best fit" \| "best fit" \| undefined` | `formatMatcher?: "basic" \| "best fit" \| "best fit" \| undefined` | `__datetimeformatoptions.formatMatcher?` | 📋 Planned | - |
| `DateTimeFormatOptions.fractionalSecondDigits?: 1 \| 2 \| 3 \| undefined` | `fractionalSecondDigits?: 1 \| 2 \| 3 \| undefined` | `__datetimeformatoptions.fractionalSecondDigits?` | 📋 Planned | - |
| `DateTimeFormatOptions.hour12?: boolean \| undefined` | `hour12?: boolean \| undefined` | `__datetimeformatoptions.hour12?` | 📋 Planned | - |
| `DateTimeFormatOptions.hour?: "numeric" \| "2-digit" \| undefined` | `hour?: "numeric" \| "2-digit" \| undefined` | `__datetimeformatoptions.hour?` | 📋 Planned | - |
| `DateTimeFormatOptions.hourCycle?: "h11" \| "h12" \| "h23" \| "h24" \| undefined` | `hourCycle?: "h11" \| "h12" \| "h23" \| "h24" \| undefined` | `__datetimeformatoptions.hourCycle?` | 📋 Planned | - |
| `DateTimeFormatOptions.localeMatcher?: "best fit" \| "lookup" \| undefined` | `localeMatcher?: "best fit" \| "lookup" \| undefined` | `__datetimeformatoptions.localeMatcher?` | 📋 Planned | - |
| `DateTimeFormatOptions.minute?: "numeric" \| "2-digit" \| undefined` | `minute?: "numeric" \| "2-digit" \| undefined` | `__datetimeformatoptions.minute?` | 📋 Planned | - |
| `DateTimeFormatOptions.month?: "numeric" \| "2-digit" \| "long" \| "short" \| "narrow" \| undefined` | `month?: "numeric" \| "2-digit" \| "long" \| "short" \| "narrow" \| undefined` | `__datetimeformatoptions.month?` | 📋 Planned | - |
| `DateTimeFormatOptions.numberingSystem?: string \| undefined` | `numberingSystem?: string \| undefined` | `__datetimeformatoptions.numberingSystem?` | 📋 Planned | - |
| `DateTimeFormatOptions.second?: "numeric" \| "2-digit" \| undefined` | `second?: "numeric" \| "2-digit" \| undefined` | `__datetimeformatoptions.second?` | 📋 Planned | - |
| `DateTimeFormatOptions.timeStyle?: "full" \| "long" \| "medium" \| "short" \| undefined` | `timeStyle?: "full" \| "long" \| "medium" \| "short" \| undefined` | `__datetimeformatoptions.timeStyle?` | 📋 Planned | - |
| `DateTimeFormatOptions.timeZone?: string \| undefined` | `timeZone?: string \| undefined` | `__datetimeformatoptions.timeZone?` | 📋 Planned | - |
| `DateTimeFormatOptions.timeZoneName?: "short" \| "long" \| "shortOffset" \| "longOffset" \| "shortGeneric" \| "longGeneric" \| undefined` | `timeZoneName?: "short" \| "long" \| "shortOffset" \| "longOffset" \| "shortGeneric" \| "longGeneric" \| undefined` | `__datetimeformatoptions.timeZoneName?` | 📋 Planned | - |
| `DateTimeFormatOptions.weekday?: "long" \| "short" \| "narrow" \| undefined` | `weekday?: "long" \| "short" \| "narrow" \| undefined` | `__datetimeformatoptions.weekday?` | 📋 Planned | - |
| `DateTimeFormatOptions.year?: "numeric" \| "2-digit" \| undefined` | `year?: "numeric" \| "2-digit" \| undefined` | `__datetimeformatoptions.year?` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `datetimeformatoptions` are organized per API under `internal/compiler/testdata/corpus/datetimeformatoptions/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/datetimeformatoptions/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
