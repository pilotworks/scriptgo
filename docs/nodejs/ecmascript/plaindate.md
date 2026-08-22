# PlainDate Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 PlainDate Specification](https://tc39.es/ecma262/#sec-plaindate-objects)  
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
| `PlainDate.add(duration: DurationLike, options?: OverflowOptions): PlainDate` | `add(duration: DurationLike, options?: OverflowOptions): PlainDate` | `__plaindate.add` | 📋 Planned | - |
| `PlainDate.compare(one: PlainDateLike, two: PlainDateLike): number` | `compare(one: PlainDateLike, two: PlainDateLike): number` | `__plaindate.compare` | 📋 Planned | - |
| `PlainDate.equals(other: PlainDateLike): boolean` | `equals(other: PlainDateLike): boolean` | `__plaindate.equals` | 📋 Planned | - |
| `PlainDate.from(item: PlainDateLike, options?: OverflowOptions): PlainDate` | `from(item: PlainDateLike, options?: OverflowOptions): PlainDate` | `__plaindate.from` | 📋 Planned | - |
| `PlainDate.readonly calendarId: string` | `readonly calendarId: string` | `__plaindate.calendarId` | 📋 Planned | - |
| `PlainDate.readonly day: number` | `readonly day: number` | `__plaindate.day` | 📋 Planned | - |
| `PlainDate.readonly dayOfWeek: number` | `readonly dayOfWeek: number` | `__plaindate.dayOfWeek` | 📋 Planned | - |
| `PlainDate.readonly dayOfYear: number` | `readonly dayOfYear: number` | `__plaindate.dayOfYear` | 📋 Planned | - |
| `PlainDate.readonly daysInMonth: number` | `readonly daysInMonth: number` | `__plaindate.daysInMonth` | 📋 Planned | - |
| `PlainDate.readonly daysInWeek: number` | `readonly daysInWeek: number` | `__plaindate.daysInWeek` | 📋 Planned | - |
| `PlainDate.readonly daysInYear: number` | `readonly daysInYear: number` | `__plaindate.daysInYear` | 📋 Planned | - |
| `PlainDate.readonly era: string \| undefined` | `readonly era: string \| undefined` | `__plaindate.era` | 📋 Planned | - |
| `PlainDate.readonly eraYear: number \| undefined` | `readonly eraYear: number \| undefined` | `__plaindate.eraYear` | 📋 Planned | - |
| `PlainDate.readonly inLeapYear: boolean` | `readonly inLeapYear: boolean` | `__plaindate.inLeapYear` | 📋 Planned | - |
| `PlainDate.readonly month: number` | `readonly month: number` | `__plaindate.month` | 📋 Planned | - |
| `PlainDate.readonly monthCode: string` | `readonly monthCode: string` | `__plaindate.monthCode` | 📋 Planned | - |
| `PlainDate.readonly monthsInYear: number` | `readonly monthsInYear: number` | `__plaindate.monthsInYear` | 📋 Planned | - |
| `PlainDate.readonly weekOfYear: number \| undefined` | `readonly weekOfYear: number \| undefined` | `__plaindate.weekOfYear` | 📋 Planned | - |
| `PlainDate.readonly year: number` | `readonly year: number` | `__plaindate.year` | 📋 Planned | - |
| `PlainDate.readonly yearOfWeek: number \| undefined` | `readonly yearOfWeek: number \| undefined` | `__plaindate.yearOfWeek` | 📋 Planned | - |
| `PlainDate.since(other: PlainDateLike, options?: RoundingOptionsWithLargestUnit<DateUnit>): Duration` | `since(other: PlainDateLike, options?: RoundingOptionsWithLargestUnit<DateUnit>): Duration` | `__plaindate.since` | 📋 Planned | - |
| `PlainDate.subtract(duration: DurationLike, options?: OverflowOptions): PlainDate` | `subtract(duration: DurationLike, options?: OverflowOptions): PlainDate` | `__plaindate.subtract` | 📋 Planned | - |
| `PlainDate.toJSON(): string` | `toJSON(): string` | `__plaindate.toJSON` | 📋 Planned | - |
| `PlainDate.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__plaindate.toLocaleString` | 📋 Planned | - |
| `PlainDate.toPlainDateTime(time?: PlainTimeLike): PlainDateTime` | `toPlainDateTime(time?: PlainTimeLike): PlainDateTime` | `__plaindate.toPlainDateTime` | 📋 Planned | - |
| `PlainDate.toPlainMonthDay(): PlainMonthDay` | `toPlainMonthDay(): PlainMonthDay` | `__plaindate.toPlainMonthDay` | 📋 Planned | - |
| `PlainDate.toPlainYearMonth(): PlainYearMonth` | `toPlainYearMonth(): PlainYearMonth` | `__plaindate.toPlainYearMonth` | 📋 Planned | - |
| `PlainDate.toString(options?: PlainDateToStringOptions): string` | `toString(options?: PlainDateToStringOptions): string` | `__plaindate.toString` | 📋 Planned | - |
| `PlainDate.toZonedDateTime(timeZone: TimeZoneLike): ZonedDateTime` | `toZonedDateTime(timeZone: TimeZoneLike): ZonedDateTime` | `__plaindate.toZonedDateTime` | 📋 Planned | - |
| `PlainDate.until(other: PlainDateLike, options?: RoundingOptionsWithLargestUnit<DateUnit>): Duration` | `until(other: PlainDateLike, options?: RoundingOptionsWithLargestUnit<DateUnit>): Duration` | `__plaindate.until` | 📋 Planned | - |
| `PlainDate.valueOf(): never` | `valueOf(): never` | `__plaindate.valueOf` | 📋 Planned | - |
| `PlainDate.with(dateLike: PartialTemporalLike<DateLikeObject>, options?: OverflowOptions): PlainDate` | `with(dateLike: PartialTemporalLike<DateLikeObject>, options?: OverflowOptions): PlainDate` | `__plaindate.with` | 📋 Planned | - |
| `PlainDate.withCalendar(calendarLike: CalendarLike): PlainDate` | `withCalendar(calendarLike: CalendarLike): PlainDate` | `__plaindate.withCalendar` | 📋 Planned | - |
| `new PlainDate(isoYear: number, isoMonth: number, isoDay: number, calendar?: string): PlainDate` | `new (isoYear: number, isoMonth: number, isoDay: number, calendar?: string): PlainDate` | `__plaindate.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `plaindate` are organized per API under `internal/compiler/testdata/corpus/plaindate/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/plaindate/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
