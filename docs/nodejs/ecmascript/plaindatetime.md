# PlainDateTime Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 PlainDateTime Specification](https://tc39.es/ecma262/#sec-plaindatetime-objects)  
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
| `PlainDateTime.add(duration: DurationLike, options?: OverflowOptions): PlainDateTime` | `add(duration: DurationLike, options?: OverflowOptions): PlainDateTime` | `__plaindatetime.add` | 📋 Planned | - |
| `PlainDateTime.compare(one: PlainDateTimeLike, two: PlainDateTimeLike): number` | `compare(one: PlainDateTimeLike, two: PlainDateTimeLike): number` | `__plaindatetime.compare` | 📋 Planned | - |
| `PlainDateTime.equals(other: PlainDateTimeLike): boolean` | `equals(other: PlainDateTimeLike): boolean` | `__plaindatetime.equals` | 📋 Planned | - |
| `PlainDateTime.from(item: PlainDateTimeLike, options?: OverflowOptions): PlainDateTime` | `from(item: PlainDateTimeLike, options?: OverflowOptions): PlainDateTime` | `__plaindatetime.from` | 📋 Planned | - |
| `PlainDateTime.readonly calendarId: string` | `readonly calendarId: string` | `__plaindatetime.calendarId` | 📋 Planned | - |
| `PlainDateTime.readonly day: number` | `readonly day: number` | `__plaindatetime.day` | 📋 Planned | - |
| `PlainDateTime.readonly dayOfWeek: number` | `readonly dayOfWeek: number` | `__plaindatetime.dayOfWeek` | 📋 Planned | - |
| `PlainDateTime.readonly dayOfYear: number` | `readonly dayOfYear: number` | `__plaindatetime.dayOfYear` | 📋 Planned | - |
| `PlainDateTime.readonly daysInMonth: number` | `readonly daysInMonth: number` | `__plaindatetime.daysInMonth` | 📋 Planned | - |
| `PlainDateTime.readonly daysInWeek: number` | `readonly daysInWeek: number` | `__plaindatetime.daysInWeek` | 📋 Planned | - |
| `PlainDateTime.readonly daysInYear: number` | `readonly daysInYear: number` | `__plaindatetime.daysInYear` | 📋 Planned | - |
| `PlainDateTime.readonly era: string \| undefined` | `readonly era: string \| undefined` | `__plaindatetime.era` | 📋 Planned | - |
| `PlainDateTime.readonly eraYear: number \| undefined` | `readonly eraYear: number \| undefined` | `__plaindatetime.eraYear` | 📋 Planned | - |
| `PlainDateTime.readonly hour: number` | `readonly hour: number` | `__plaindatetime.hour` | 📋 Planned | - |
| `PlainDateTime.readonly inLeapYear: boolean` | `readonly inLeapYear: boolean` | `__plaindatetime.inLeapYear` | 📋 Planned | - |
| `PlainDateTime.readonly microsecond: number` | `readonly microsecond: number` | `__plaindatetime.microsecond` | 📋 Planned | - |
| `PlainDateTime.readonly millisecond: number` | `readonly millisecond: number` | `__plaindatetime.millisecond` | 📋 Planned | - |
| `PlainDateTime.readonly minute: number` | `readonly minute: number` | `__plaindatetime.minute` | 📋 Planned | - |
| `PlainDateTime.readonly month: number` | `readonly month: number` | `__plaindatetime.month` | 📋 Planned | - |
| `PlainDateTime.readonly monthCode: string` | `readonly monthCode: string` | `__plaindatetime.monthCode` | 📋 Planned | - |
| `PlainDateTime.readonly monthsInYear: number` | `readonly monthsInYear: number` | `__plaindatetime.monthsInYear` | 📋 Planned | - |
| `PlainDateTime.readonly nanosecond: number` | `readonly nanosecond: number` | `__plaindatetime.nanosecond` | 📋 Planned | - |
| `PlainDateTime.readonly second: number` | `readonly second: number` | `__plaindatetime.second` | 📋 Planned | - |
| `PlainDateTime.readonly weekOfYear: number \| undefined` | `readonly weekOfYear: number \| undefined` | `__plaindatetime.weekOfYear` | 📋 Planned | - |
| `PlainDateTime.readonly year: number` | `readonly year: number` | `__plaindatetime.year` | 📋 Planned | - |
| `PlainDateTime.readonly yearOfWeek: number \| undefined` | `readonly yearOfWeek: number \| undefined` | `__plaindatetime.yearOfWeek` | 📋 Planned | - |
| `PlainDateTime.round(roundTo: PluralizeUnit<"day" \| TimeUnit>): PlainDateTime` | `round(roundTo: PluralizeUnit<"day" \| TimeUnit>): PlainDateTime` | `__plaindatetime.round` | 📋 Planned | - |
| `PlainDateTime.since(other: PlainDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `since(other: PlainDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `__plaindatetime.since` | 📋 Planned | - |
| `PlainDateTime.subtract(duration: DurationLike, options?: OverflowOptions): PlainDateTime` | `subtract(duration: DurationLike, options?: OverflowOptions): PlainDateTime` | `__plaindatetime.subtract` | 📋 Planned | - |
| `PlainDateTime.toJSON(): string` | `toJSON(): string` | `__plaindatetime.toJSON` | 📋 Planned | - |
| `PlainDateTime.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__plaindatetime.toLocaleString` | 📋 Planned | - |
| `PlainDateTime.toPlainDate(): PlainDate` | `toPlainDate(): PlainDate` | `__plaindatetime.toPlainDate` | 📋 Planned | - |
| `PlainDateTime.toPlainTime(): PlainTime` | `toPlainTime(): PlainTime` | `__plaindatetime.toPlainTime` | 📋 Planned | - |
| `PlainDateTime.toString(options?: PlainDateTimeToStringOptions): string` | `toString(options?: PlainDateTimeToStringOptions): string` | `__plaindatetime.toString` | 📋 Planned | - |
| `PlainDateTime.toZonedDateTime(timeZone: TimeZoneLike, options?: DisambiguationOptions): ZonedDateTime` | `toZonedDateTime(timeZone: TimeZoneLike, options?: DisambiguationOptions): ZonedDateTime` | `__plaindatetime.toZonedDateTime` | 📋 Planned | - |
| `PlainDateTime.until(other: PlainDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `until(other: PlainDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `__plaindatetime.until` | 📋 Planned | - |
| `PlainDateTime.valueOf(): never` | `valueOf(): never` | `__plaindatetime.valueOf` | 📋 Planned | - |
| `PlainDateTime.with(dateTimeLike: PartialTemporalLike<DateTimeLikeObject>, options?: OverflowOptions): PlainDateTime` | `with(dateTimeLike: PartialTemporalLike<DateTimeLikeObject>, options?: OverflowOptions): PlainDateTime` | `__plaindatetime.with` | 📋 Planned | - |
| `PlainDateTime.withCalendar(calendar: CalendarLike): PlainDateTime` | `withCalendar(calendar: CalendarLike): PlainDateTime` | `__plaindatetime.withCalendar` | 📋 Planned | - |
| `PlainDateTime.withPlainTime(plainTime?: PlainTimeLike): PlainDateTime` | `withPlainTime(plainTime?: PlainTimeLike): PlainDateTime` | `__plaindatetime.withPlainTime` | 📋 Planned | - |
| `new PlainDateTime(isoYear: number, isoMonth: number, isoDay: number, hour?: number, minute?: number, second?: number, millisecond?: number, microsecond?: number, nanosecond?: number, calendar?: string): PlainDateTime` | `new (isoYear: number, isoMonth: number, isoDay: number, hour?: number, minute?: number, second?: number, millisecond?: number, microsecond?: number, nanosecond?: number, calendar?: string): PlainDateTime` | `__plaindatetime.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `plaindatetime` are organized per API under `internal/compiler/testdata/corpus/plaindatetime/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/plaindatetime/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
