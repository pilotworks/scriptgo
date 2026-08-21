# PlainTime Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 PlainTime Specification](https://tc39.es/ecma262/#sec-plaintime-objects)  
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
| `PlainTime.add(duration: DurationLike): PlainTime` | `add(duration: DurationLike): PlainTime` | `__plaintime.add` | 📋 Planned | - |
| `PlainTime.compare(one: PlainTimeLike, two: PlainTimeLike): number` | `compare(one: PlainTimeLike, two: PlainTimeLike): number` | `__plaintime.compare` | 📋 Planned | - |
| `PlainTime.equals(other: PlainTimeLike): boolean` | `equals(other: PlainTimeLike): boolean` | `__plaintime.equals` | 📋 Planned | - |
| `PlainTime.from(item: PlainTimeLike, options?: OverflowOptions): PlainTime` | `from(item: PlainTimeLike, options?: OverflowOptions): PlainTime` | `__plaintime.from` | 📋 Planned | - |
| `PlainTime.new (hour?: number, minute?: number, second?: number, millisecond?: number, microsecond?: number, nanosecond?: number): PlainTime` | `new (hour?: number, minute?: number, second?: number, millisecond?: number, microsecond?: number, nanosecond?: number): PlainTime` | `__plaintime.new` | 📋 Planned | - |
| `PlainTime.readonly hour: number` | `readonly hour: number` | `__plaintime.hour` | 📋 Planned | - |
| `PlainTime.readonly microsecond: number` | `readonly microsecond: number` | `__plaintime.microsecond` | 📋 Planned | - |
| `PlainTime.readonly millisecond: number` | `readonly millisecond: number` | `__plaintime.millisecond` | 📋 Planned | - |
| `PlainTime.readonly minute: number` | `readonly minute: number` | `__plaintime.minute` | 📋 Planned | - |
| `PlainTime.readonly nanosecond: number` | `readonly nanosecond: number` | `__plaintime.nanosecond` | 📋 Planned | - |
| `PlainTime.readonly prototype: PlainTime` | `readonly prototype: PlainTime` | `__plaintime.prototype` | 📋 Planned | - |
| `PlainTime.readonly second: number` | `readonly second: number` | `__plaintime.second` | 📋 Planned | - |
| `PlainTime.round(roundTo: PluralizeUnit<TimeUnit>): PlainTime` | `round(roundTo: PluralizeUnit<TimeUnit>): PlainTime` | `__plaintime.round` | 📋 Planned | - |
| `PlainTime.since(other: PlainTimeLike, options?: RoundingOptionsWithLargestUnit<TimeUnit>): Duration` | `since(other: PlainTimeLike, options?: RoundingOptionsWithLargestUnit<TimeUnit>): Duration` | `__plaintime.since` | 📋 Planned | - |
| `PlainTime.subtract(duration: DurationLike): PlainTime` | `subtract(duration: DurationLike): PlainTime` | `__plaintime.subtract` | 📋 Planned | - |
| `PlainTime.toJSON(): string` | `toJSON(): string` | `__plaintime.toJSON` | 📋 Planned | - |
| `PlainTime.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__plaintime.toLocaleString` | 📋 Planned | - |
| `PlainTime.toString(options?: PlainTimeToStringOptions): string` | `toString(options?: PlainTimeToStringOptions): string` | `__plaintime.toString` | 📋 Planned | - |
| `PlainTime.until(other: PlainTimeLike, options?: RoundingOptionsWithLargestUnit<TimeUnit>): Duration` | `until(other: PlainTimeLike, options?: RoundingOptionsWithLargestUnit<TimeUnit>): Duration` | `__plaintime.until` | 📋 Planned | - |
| `PlainTime.valueOf(): never` | `valueOf(): never` | `__plaintime.valueOf` | 📋 Planned | - |
| `PlainTime.with(timeLike: PartialTemporalLike<TimeLikeObject>, options?: OverflowOptions): PlainTime` | `with(timeLike: PartialTemporalLike<TimeLikeObject>, options?: OverflowOptions): PlainTime` | `__plaintime.with` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `plaintime` are organized per API under `internal/compiler/testdata/corpus/plaintime/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/plaintime/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
