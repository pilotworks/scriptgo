# NumberFormatPartTypeRegistry Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 NumberFormatPartTypeRegistry Specification](https://tc39.es/ecma262/#sec-numberformatparttyperegistry-objects)  
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
| `NumberFormatPartTypeRegistry.compact: never` | `compact: never` | `__numberformatparttyperegistry.compact` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.currency: never` | `currency: never` | `__numberformatparttyperegistry.currency` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.decimal: never` | `decimal: never` | `__numberformatparttyperegistry.decimal` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.exponentInteger: never` | `exponentInteger: never` | `__numberformatparttyperegistry.exponentInteger` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.exponentMinusSign: never` | `exponentMinusSign: never` | `__numberformatparttyperegistry.exponentMinusSign` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.exponentSeparator: never` | `exponentSeparator: never` | `__numberformatparttyperegistry.exponentSeparator` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.fraction: never` | `fraction: never` | `__numberformatparttyperegistry.fraction` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.group: never` | `group: never` | `__numberformatparttyperegistry.group` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.infinity: never` | `infinity: never` | `__numberformatparttyperegistry.infinity` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.integer: never` | `integer: never` | `__numberformatparttyperegistry.integer` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.literal: never` | `literal: never` | `__numberformatparttyperegistry.literal` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.minusSign: never` | `minusSign: never` | `__numberformatparttyperegistry.minusSign` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.nan: never` | `nan: never` | `__numberformatparttyperegistry.nan` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.percent: never` | `percent: never` | `__numberformatparttyperegistry.percent` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.percentSign: never` | `percentSign: never` | `__numberformatparttyperegistry.percentSign` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.plusSign: never` | `plusSign: never` | `__numberformatparttyperegistry.plusSign` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.unit: never` | `unit: never` | `__numberformatparttyperegistry.unit` | 📋 Planned | - |
| `NumberFormatPartTypeRegistry.unknown: never` | `unknown: never` | `__numberformatparttyperegistry.unknown` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `numberformatparttyperegistry` are organized per API under `internal/compiler/testdata/corpus/numberformatparttyperegistry/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/numberformatparttyperegistry/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
