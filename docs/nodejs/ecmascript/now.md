# Now Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Now Specification](https://tc39.es/ecma262/#sec-now-objects)  
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
| `Now.instant(): Instant` | `instant(): Instant` | `__now.instant` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |
| `Now.plainDateISO(timeZone?: TimeZoneLike): PlainDate` | `plainDateISO(timeZone?: TimeZoneLike): PlainDate` | `__now.plainDateISO` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |
| `Now.plainDateTimeISO(timeZone?: TimeZoneLike): PlainDateTime` | `plainDateTimeISO(timeZone?: TimeZoneLike): PlainDateTime` | `__now.plainDateTimeISO` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |
| `Now.plainTimeISO(timeZone?: TimeZoneLike): PlainTime` | `plainTimeISO(timeZone?: TimeZoneLike): PlainTime` | `__now.plainTimeISO` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |
| `Now.timeZoneId(): string` | `timeZoneId(): string` | `__now.timeZoneId` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |
| `Now.zonedDateTimeISO(timeZone?: TimeZoneLike): ZonedDateTime` | `zonedDateTimeISO(timeZone?: TimeZoneLike): ZonedDateTime` | `__now.zonedDateTimeISO` | ✅ Done | `internal/compiler/testdata/corpus/api/now.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `now` are organized per API under `internal/compiler/testdata/corpus/now/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/now/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
