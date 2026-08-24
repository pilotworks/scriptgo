# DateTimeFormatPartTypesRegistry Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 DateTimeFormatPartTypesRegistry Specification](https://tc39.es/ecma262/#sec-datetimeformatparttypesregistry-objects)  
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
| `DateTimeFormatPartTypesRegistry.day: any` | `day: any` | `__datetimeformatparttypesregistry.day` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.dayPeriod: any` | `dayPeriod: any` | `__datetimeformatparttypesregistry.dayPeriod` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.era: any` | `era: any` | `__datetimeformatparttypesregistry.era` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.fractionalSecond: any` | `fractionalSecond: any` | `__datetimeformatparttypesregistry.fractionalSecond` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.hour: any` | `hour: any` | `__datetimeformatparttypesregistry.hour` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.literal: any` | `literal: any` | `__datetimeformatparttypesregistry.literal` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.minute: any` | `minute: any` | `__datetimeformatparttypesregistry.minute` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.month: any` | `month: any` | `__datetimeformatparttypesregistry.month` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.second: any` | `second: any` | `__datetimeformatparttypesregistry.second` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.timeZoneName: any` | `timeZoneName: any` | `__datetimeformatparttypesregistry.timeZoneName` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.unknown: never` | `unknown: never` | `__datetimeformatparttypesregistry.unknown` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.weekday: any` | `weekday: any` | `__datetimeformatparttypesregistry.weekday` | 📋 Planned | - |
| `DateTimeFormatPartTypesRegistry.year: any` | `year: any` | `__datetimeformatparttypesregistry.year` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `datetimeformatparttypesregistry` are organized per API under `internal/compiler/testdata/corpus/datetimeformatparttypesregistry/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/datetimeformatparttypesregistry/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
