# PlainMonthDay Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 PlainMonthDay Specification](https://tc39.es/ecma262/#sec-plainmonthday-objects)  
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
| `PlainMonthDay.equals(other: PlainMonthDayLike): boolean` | `equals(other: PlainMonthDayLike): boolean` | `__plainmonthday.equals` | 📋 Planned | - |
| `PlainMonthDay.from(item: PlainMonthDayLike, options?: OverflowOptions): PlainMonthDay` | `from(item: PlainMonthDayLike, options?: OverflowOptions): PlainMonthDay` | `__plainmonthday.from` | 📋 Planned | - |
| `PlainMonthDay.readonly calendarId: string` | `readonly calendarId: string` | `__plainmonthday.calendarId` | 📋 Planned | - |
| `PlainMonthDay.readonly day: number` | `readonly day: number` | `__plainmonthday.day` | 📋 Planned | - |
| `PlainMonthDay.readonly monthCode: string` | `readonly monthCode: string` | `__plainmonthday.monthCode` | 📋 Planned | - |
| `PlainMonthDay.toJSON(): string` | `toJSON(): string` | `__plainmonthday.toJSON` | 📋 Planned | - |
| `PlainMonthDay.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__plainmonthday.toLocaleString` | 📋 Planned | - |
| `PlainMonthDay.toPlainDate(item: PlainMonthDayToPlainDateOptions): PlainDate` | `toPlainDate(item: PlainMonthDayToPlainDateOptions): PlainDate` | `__plainmonthday.toPlainDate` | 📋 Planned | - |
| `PlainMonthDay.toString(options?: PlainDateToStringOptions): string` | `toString(options?: PlainDateToStringOptions): string` | `__plainmonthday.toString` | 📋 Planned | - |
| `PlainMonthDay.valueOf(): never` | `valueOf(): never` | `__plainmonthday.valueOf` | 📋 Planned | - |
| `PlainMonthDay.with(monthDayLike: PartialTemporalLike<DateLikeObject>, options?: OverflowOptions): PlainMonthDay` | `with(monthDayLike: PartialTemporalLike<DateLikeObject>, options?: OverflowOptions): PlainMonthDay` | `__plainmonthday.with` | 📋 Planned | - |
| `new PlainMonthDay(isoMonth: number, isoDay: number, calendar?: string, referenceISOYear?: number): PlainMonthDay` | `new (isoMonth: number, isoDay: number, calendar?: string, referenceISOYear?: number): PlainMonthDay` | `__plainmonthday.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `plainmonthday` are organized per API under `internal/compiler/testdata/corpus/plainmonthday/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/plainmonthday/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
