# ZonedDateTime Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 ZonedDateTime Specification](https://tc39.es/ecma262/#sec-zoneddatetime-objects)  
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
| `ZonedDateTime.add(duration: DurationLike, options?: OverflowOptions): ZonedDateTime` | `add(duration: DurationLike, options?: OverflowOptions): ZonedDateTime` | `__zoneddatetime.add` | 📋 Planned | - |
| `ZonedDateTime.compare(one: ZonedDateTimeLike, two: ZonedDateTimeLike): number` | `compare(one: ZonedDateTimeLike, two: ZonedDateTimeLike): number` | `__zoneddatetime.compare` | 📋 Planned | - |
| `ZonedDateTime.equals(other: ZonedDateTimeLike): boolean` | `equals(other: ZonedDateTimeLike): boolean` | `__zoneddatetime.equals` | 📋 Planned | - |
| `ZonedDateTime.from(item: ZonedDateTimeLike, options?: ZonedDateTimeFromOptions): ZonedDateTime` | `from(item: ZonedDateTimeLike, options?: ZonedDateTimeFromOptions): ZonedDateTime` | `__zoneddatetime.from` | 📋 Planned | - |
| `ZonedDateTime.getTimeZoneTransition(direction: "next" \| "previous"): ZonedDateTime \| null` | `getTimeZoneTransition(direction: "next" \| "previous"): ZonedDateTime \| null` | `__zoneddatetime.getTimeZoneTransition` | 📋 Planned | - |
| `ZonedDateTime.new (epochNanoseconds: bigint, timeZone: string, calendar?: string): ZonedDateTime` | `new (epochNanoseconds: bigint, timeZone: string, calendar?: string): ZonedDateTime` | `__zoneddatetime.new` | 📋 Planned | - |
| `ZonedDateTime.readonly calendarId: string` | `readonly calendarId: string` | `__zoneddatetime.calendarId` | 📋 Planned | - |
| `ZonedDateTime.readonly day: number` | `readonly day: number` | `__zoneddatetime.day` | 📋 Planned | - |
| `ZonedDateTime.readonly dayOfWeek: number` | `readonly dayOfWeek: number` | `__zoneddatetime.dayOfWeek` | 📋 Planned | - |
| `ZonedDateTime.readonly dayOfYear: number` | `readonly dayOfYear: number` | `__zoneddatetime.dayOfYear` | 📋 Planned | - |
| `ZonedDateTime.readonly daysInMonth: number` | `readonly daysInMonth: number` | `__zoneddatetime.daysInMonth` | 📋 Planned | - |
| `ZonedDateTime.readonly daysInWeek: number` | `readonly daysInWeek: number` | `__zoneddatetime.daysInWeek` | 📋 Planned | - |
| `ZonedDateTime.readonly daysInYear: number` | `readonly daysInYear: number` | `__zoneddatetime.daysInYear` | 📋 Planned | - |
| `ZonedDateTime.readonly epochMilliseconds: number` | `readonly epochMilliseconds: number` | `__zoneddatetime.epochMilliseconds` | 📋 Planned | - |
| `ZonedDateTime.readonly epochNanoseconds: bigint` | `readonly epochNanoseconds: bigint` | `__zoneddatetime.epochNanoseconds` | 📋 Planned | - |
| `ZonedDateTime.readonly era: string \| undefined` | `readonly era: string \| undefined` | `__zoneddatetime.era` | 📋 Planned | - |
| `ZonedDateTime.readonly eraYear: number \| undefined` | `readonly eraYear: number \| undefined` | `__zoneddatetime.eraYear` | 📋 Planned | - |
| `ZonedDateTime.readonly hour: number` | `readonly hour: number` | `__zoneddatetime.hour` | 📋 Planned | - |
| `ZonedDateTime.readonly hoursInDay: number` | `readonly hoursInDay: number` | `__zoneddatetime.hoursInDay` | 📋 Planned | - |
| `ZonedDateTime.readonly inLeapYear: boolean` | `readonly inLeapYear: boolean` | `__zoneddatetime.inLeapYear` | 📋 Planned | - |
| `ZonedDateTime.readonly microsecond: number` | `readonly microsecond: number` | `__zoneddatetime.microsecond` | 📋 Planned | - |
| `ZonedDateTime.readonly millisecond: number` | `readonly millisecond: number` | `__zoneddatetime.millisecond` | 📋 Planned | - |
| `ZonedDateTime.readonly minute: number` | `readonly minute: number` | `__zoneddatetime.minute` | 📋 Planned | - |
| `ZonedDateTime.readonly month: number` | `readonly month: number` | `__zoneddatetime.month` | 📋 Planned | - |
| `ZonedDateTime.readonly monthCode: string` | `readonly monthCode: string` | `__zoneddatetime.monthCode` | 📋 Planned | - |
| `ZonedDateTime.readonly monthsInYear: number` | `readonly monthsInYear: number` | `__zoneddatetime.monthsInYear` | 📋 Planned | - |
| `ZonedDateTime.readonly nanosecond: number` | `readonly nanosecond: number` | `__zoneddatetime.nanosecond` | 📋 Planned | - |
| `ZonedDateTime.readonly offset: string` | `readonly offset: string` | `__zoneddatetime.offset` | 📋 Planned | - |
| `ZonedDateTime.readonly offsetNanoseconds: number` | `readonly offsetNanoseconds: number` | `__zoneddatetime.offsetNanoseconds` | 📋 Planned | - |
| `ZonedDateTime.readonly prototype: ZonedDateTime` | `readonly prototype: ZonedDateTime` | `__zoneddatetime.prototype` | 📋 Planned | - |
| `ZonedDateTime.readonly second: number` | `readonly second: number` | `__zoneddatetime.second` | 📋 Planned | - |
| `ZonedDateTime.readonly timeZoneId: string` | `readonly timeZoneId: string` | `__zoneddatetime.timeZoneId` | 📋 Planned | - |
| `ZonedDateTime.readonly weekOfYear: number \| undefined` | `readonly weekOfYear: number \| undefined` | `__zoneddatetime.weekOfYear` | 📋 Planned | - |
| `ZonedDateTime.readonly year: number` | `readonly year: number` | `__zoneddatetime.year` | 📋 Planned | - |
| `ZonedDateTime.readonly yearOfWeek: number \| undefined` | `readonly yearOfWeek: number \| undefined` | `__zoneddatetime.yearOfWeek` | 📋 Planned | - |
| `ZonedDateTime.round(roundTo: PluralizeUnit<"day" \| TimeUnit>): ZonedDateTime` | `round(roundTo: PluralizeUnit<"day" \| TimeUnit>): ZonedDateTime` | `__zoneddatetime.round` | 📋 Planned | - |
| `ZonedDateTime.since(other: ZonedDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `since(other: ZonedDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `__zoneddatetime.since` | 📋 Planned | - |
| `ZonedDateTime.startOfDay(): ZonedDateTime` | `startOfDay(): ZonedDateTime` | `__zoneddatetime.startOfDay` | 📋 Planned | - |
| `ZonedDateTime.subtract(duration: DurationLike, options?: OverflowOptions): ZonedDateTime` | `subtract(duration: DurationLike, options?: OverflowOptions): ZonedDateTime` | `__zoneddatetime.subtract` | 📋 Planned | - |
| `ZonedDateTime.toInstant(): Instant` | `toInstant(): Instant` | `__zoneddatetime.toInstant` | 📋 Planned | - |
| `ZonedDateTime.toJSON(): string` | `toJSON(): string` | `__zoneddatetime.toJSON` | 📋 Planned | - |
| `ZonedDateTime.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__zoneddatetime.toLocaleString` | 📋 Planned | - |
| `ZonedDateTime.toPlainDate(): PlainDate` | `toPlainDate(): PlainDate` | `__zoneddatetime.toPlainDate` | 📋 Planned | - |
| `ZonedDateTime.toPlainDateTime(): PlainDateTime` | `toPlainDateTime(): PlainDateTime` | `__zoneddatetime.toPlainDateTime` | 📋 Planned | - |
| `ZonedDateTime.toPlainTime(): PlainTime` | `toPlainTime(): PlainTime` | `__zoneddatetime.toPlainTime` | 📋 Planned | - |
| `ZonedDateTime.toString(options?: ZonedDateTimeToStringOptions): string` | `toString(options?: ZonedDateTimeToStringOptions): string` | `__zoneddatetime.toString` | 📋 Planned | - |
| `ZonedDateTime.until(other: ZonedDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `until(other: ZonedDateTimeLike, options?: RoundingOptionsWithLargestUnit<DateUnit \| TimeUnit>): Duration` | `__zoneddatetime.until` | 📋 Planned | - |
| `ZonedDateTime.valueOf(): never` | `valueOf(): never` | `__zoneddatetime.valueOf` | 📋 Planned | - |
| `ZonedDateTime.with(zonedDateTimeLike: PartialTemporalLike<ZonedDateTimeLikeObject>, options?: ZonedDateTimeFromOptions): ZonedDateTime` | `with(zonedDateTimeLike: PartialTemporalLike<ZonedDateTimeLikeObject>, options?: ZonedDateTimeFromOptions): ZonedDateTime` | `__zoneddatetime.with` | 📋 Planned | - |
| `ZonedDateTime.withCalendar(calendar: CalendarLike): ZonedDateTime` | `withCalendar(calendar: CalendarLike): ZonedDateTime` | `__zoneddatetime.withCalendar` | 📋 Planned | - |
| `ZonedDateTime.withPlainTime(plainTime?: PlainTimeLike): ZonedDateTime` | `withPlainTime(plainTime?: PlainTimeLike): ZonedDateTime` | `__zoneddatetime.withPlainTime` | 📋 Planned | - |
| `ZonedDateTime.withTimeZone(timeZone: TimeZoneLike): ZonedDateTime` | `withTimeZone(timeZone: TimeZoneLike): ZonedDateTime` | `__zoneddatetime.withTimeZone` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `zoneddatetime` are organized per API under `internal/compiler/testdata/corpus/zoneddatetime/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/zoneddatetime/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
