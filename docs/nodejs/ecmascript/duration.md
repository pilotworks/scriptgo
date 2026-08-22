# Duration Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Duration Specification](https://tc39.es/ecma262/#sec-duration-objects)  
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
| `Duration.abs(): Duration` | `abs(): Duration` | `__duration.abs` | 📋 Planned | - |
| `Duration.add(other: DurationLike): Duration` | `add(other: DurationLike): Duration` | `__duration.add` | 📋 Planned | - |
| `Duration.compare(one: DurationLike, two: DurationLike, options?: DurationRelativeToOptions): number` | `compare(one: DurationLike, two: DurationLike, options?: DurationRelativeToOptions): number` | `__duration.compare` | 📋 Planned | - |
| `Duration.from(item: DurationLike): Duration` | `from(item: DurationLike): Duration` | `__duration.from` | 📋 Planned | - |
| `Duration.negated(): Duration` | `negated(): Duration` | `__duration.negated` | 📋 Planned | - |
| `Duration.readonly blank: boolean` | `readonly blank: boolean` | `__duration.blank` | 📋 Planned | - |
| `Duration.readonly days: number` | `readonly days: number` | `__duration.days` | 📋 Planned | - |
| `Duration.readonly hours: number` | `readonly hours: number` | `__duration.hours` | 📋 Planned | - |
| `Duration.readonly microseconds: number` | `readonly microseconds: number` | `__duration.microseconds` | 📋 Planned | - |
| `Duration.readonly milliseconds: number` | `readonly milliseconds: number` | `__duration.milliseconds` | 📋 Planned | - |
| `Duration.readonly minutes: number` | `readonly minutes: number` | `__duration.minutes` | 📋 Planned | - |
| `Duration.readonly months: number` | `readonly months: number` | `__duration.months` | 📋 Planned | - |
| `Duration.readonly nanoseconds: number` | `readonly nanoseconds: number` | `__duration.nanoseconds` | 📋 Planned | - |
| `Duration.readonly seconds: number` | `readonly seconds: number` | `__duration.seconds` | 📋 Planned | - |
| `Duration.readonly sign: number` | `readonly sign: number` | `__duration.sign` | 📋 Planned | - |
| `Duration.readonly weeks: number` | `readonly weeks: number` | `__duration.weeks` | 📋 Planned | - |
| `Duration.readonly years: number` | `readonly years: number` | `__duration.years` | 📋 Planned | - |
| `Duration.round(roundTo: PluralizeUnit<"day" \| TimeUnit>): Duration` | `round(roundTo: PluralizeUnit<"day" \| TimeUnit>): Duration` | `__duration.round` | 📋 Planned | - |
| `Duration.subtract(other: DurationLike): Duration` | `subtract(other: DurationLike): Duration` | `__duration.subtract` | 📋 Planned | - |
| `Duration.toJSON(): string` | `toJSON(): string` | `__duration.toJSON` | 📋 Planned | - |
| `Duration.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DurationFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DurationFormatOptions): string` | `__duration.toLocaleString` | 📋 Planned | - |
| `Duration.toString(options?: DurationToStringOptions): string` | `toString(options?: DurationToStringOptions): string` | `__duration.toString` | 📋 Planned | - |
| `Duration.total(totalOf: PluralizeUnit<"day" \| TimeUnit>): number` | `total(totalOf: PluralizeUnit<"day" \| TimeUnit>): number` | `__duration.total` | 📋 Planned | - |
| `Duration.valueOf(): never` | `valueOf(): never` | `__duration.valueOf` | 📋 Planned | - |
| `Duration.with(durationLike: PartialTemporalLike<DurationLikeObject>): Duration` | `with(durationLike: PartialTemporalLike<DurationLikeObject>): Duration` | `__duration.with` | 📋 Planned | - |
| `new Duration(years?: number, months?: number, weeks?: number, days?: number, hours?: number, minutes?: number, seconds?: number, milliseconds?: number, microseconds?: number, nanoseconds?: number): Duration` | `new (years?: number, months?: number, weeks?: number, days?: number, hours?: number, minutes?: number, seconds?: number, milliseconds?: number, microseconds?: number, nanoseconds?: number): Duration` | `__duration.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `duration` are organized per API under `internal/compiler/testdata/corpus/duration/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/duration/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
