# PlainYearMonth Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 PlainYearMonth Specification](https://tc39.es/ecma262/#sec-plainyearmonth-objects)  
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
| `PlainYearMonth.add(duration: DurationLike, options?: OverflowOptions): PlainYearMonth` | `add(duration: DurationLike, options?: OverflowOptions): PlainYearMonth` | `__plainyearmonth.add` | 📋 Planned | - |
| `PlainYearMonth.compare(one: PlainYearMonthLike, two: PlainYearMonthLike): number` | `compare(one: PlainYearMonthLike, two: PlainYearMonthLike): number` | `__plainyearmonth.compare` | 📋 Planned | - |
| `PlainYearMonth.equals(other: PlainYearMonthLike): boolean` | `equals(other: PlainYearMonthLike): boolean` | `__plainyearmonth.equals` | 📋 Planned | - |
| `PlainYearMonth.from(item: PlainYearMonthLike, options?: OverflowOptions): PlainYearMonth` | `from(item: PlainYearMonthLike, options?: OverflowOptions): PlainYearMonth` | `__plainyearmonth.from` | 📋 Planned | - |
| `PlainYearMonth.readonly calendarId: string` | `readonly calendarId: string` | `__plainyearmonth.calendarId` | 📋 Planned | - |
| `PlainYearMonth.readonly daysInMonth: number` | `readonly daysInMonth: number` | `__plainyearmonth.daysInMonth` | 📋 Planned | - |
| `PlainYearMonth.readonly daysInYear: number` | `readonly daysInYear: number` | `__plainyearmonth.daysInYear` | 📋 Planned | - |
| `PlainYearMonth.readonly era: string \| undefined` | `readonly era: string \| undefined` | `__plainyearmonth.era` | 📋 Planned | - |
| `PlainYearMonth.readonly eraYear: number \| undefined` | `readonly eraYear: number \| undefined` | `__plainyearmonth.eraYear` | 📋 Planned | - |
| `PlainYearMonth.readonly inLeapYear: boolean` | `readonly inLeapYear: boolean` | `__plainyearmonth.inLeapYear` | 📋 Planned | - |
| `PlainYearMonth.readonly month: number` | `readonly month: number` | `__plainyearmonth.month` | 📋 Planned | - |
| `PlainYearMonth.readonly monthCode: string` | `readonly monthCode: string` | `__plainyearmonth.monthCode` | 📋 Planned | - |
| `PlainYearMonth.readonly monthsInYear: number` | `readonly monthsInYear: number` | `__plainyearmonth.monthsInYear` | 📋 Planned | - |
| `PlainYearMonth.readonly year: number` | `readonly year: number` | `__plainyearmonth.year` | 📋 Planned | - |
| `PlainYearMonth.since(other: PlainYearMonthLike, options?: RoundingOptionsWithLargestUnit<"year" \| "month">): Duration` | `since(other: PlainYearMonthLike, options?: RoundingOptionsWithLargestUnit<"year" \| "month">): Duration` | `__plainyearmonth.since` | 📋 Planned | - |
| `PlainYearMonth.subtract(duration: DurationLike, options?: OverflowOptions): PlainYearMonth` | `subtract(duration: DurationLike, options?: OverflowOptions): PlainYearMonth` | `__plainyearmonth.subtract` | 📋 Planned | - |
| `PlainYearMonth.toJSON(): string` | `toJSON(): string` | `__plainyearmonth.toJSON` | 📋 Planned | - |
| `PlainYearMonth.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__plainyearmonth.toLocaleString` | 📋 Planned | - |
| `PlainYearMonth.toPlainDate(item: PlainYearMonthToPlainDateOptions): PlainDate` | `toPlainDate(item: PlainYearMonthToPlainDateOptions): PlainDate` | `__plainyearmonth.toPlainDate` | 📋 Planned | - |
| `PlainYearMonth.toString(options?: PlainDateToStringOptions): string` | `toString(options?: PlainDateToStringOptions): string` | `__plainyearmonth.toString` | 📋 Planned | - |
| `PlainYearMonth.until(other: PlainYearMonthLike, options?: RoundingOptionsWithLargestUnit<"year" \| "month">): Duration` | `until(other: PlainYearMonthLike, options?: RoundingOptionsWithLargestUnit<"year" \| "month">): Duration` | `__plainyearmonth.until` | 📋 Planned | - |
| `PlainYearMonth.valueOf(): never` | `valueOf(): never` | `__plainyearmonth.valueOf` | 📋 Planned | - |
| `PlainYearMonth.with(yearMonthLike: PartialTemporalLike<YearMonthLikeObject>, options?: OverflowOptions): PlainYearMonth` | `with(yearMonthLike: PartialTemporalLike<YearMonthLikeObject>, options?: OverflowOptions): PlainYearMonth` | `__plainyearmonth.with` | 📋 Planned | - |
| `new PlainYearMonth(isoYear: number, isoMonth: number, calendar?: string, referenceISODay?: number): PlainYearMonth` | `new (isoYear: number, isoMonth: number, calendar?: string, referenceISODay?: number): PlainYearMonth` | `__plainyearmonth.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `plainyearmonth` are organized per API under `internal/compiler/testdata/corpus/plainyearmonth/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/plainyearmonth/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
