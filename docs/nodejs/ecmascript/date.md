# Date Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Date Specification](https://tc39.es/ecma262/#sec-date-objects)  
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
| `Date.UTC(year: number, monthIndex?: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): number` | `UTC(year: number, monthIndex?: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): number` | `__date.UTC` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getDate(): number` | `getDate(): number` | `__date.getDate` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getDay(): number` | `getDay(): number` | `__date.getDay` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getFullYear(): number` | `getFullYear(): number` | `__date.getFullYear` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getHours(): number` | `getHours(): number` | `__date.getHours` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getMilliseconds(): number` | `getMilliseconds(): number` | `__date.getMilliseconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getMinutes(): number` | `getMinutes(): number` | `__date.getMinutes` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getMonth(): number` | `getMonth(): number` | `__date.getMonth` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getSeconds(): number` | `getSeconds(): number` | `__date.getSeconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getTime(): number` | `getTime(): number` | `__date.getTime` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getTimezoneOffset(): number` | `getTimezoneOffset(): number` | `__date.getTimezoneOffset` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCDate(): number` | `getUTCDate(): number` | `__date.getUTCDate` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCDay(): number` | `getUTCDay(): number` | `__date.getUTCDay` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCFullYear(): number` | `getUTCFullYear(): number` | `__date.getUTCFullYear` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCHours(): number` | `getUTCHours(): number` | `__date.getUTCHours` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCMilliseconds(): number` | `getUTCMilliseconds(): number` | `__date.getUTCMilliseconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCMinutes(): number` | `getUTCMinutes(): number` | `__date.getUTCMinutes` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCMonth(): number` | `getUTCMonth(): number` | `__date.getUTCMonth` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.getUTCSeconds(): number` | `getUTCSeconds(): number` | `__date.getUTCSeconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.now(): number` | `now(): number` | `__date.now` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.parse(s: string): number` | `parse(s: string): number` | `__date.parse` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setDate(date: number): number` | `setDate(date: number): number` | `__date.setDate` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setFullYear(year: number, month?: number, date?: number): number` | `setFullYear(year: number, month?: number, date?: number): number` | `__date.setFullYear` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setHours(hours: number, min?: number, sec?: number, ms?: number): number` | `setHours(hours: number, min?: number, sec?: number, ms?: number): number` | `__date.setHours` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setMilliseconds(ms: number): number` | `setMilliseconds(ms: number): number` | `__date.setMilliseconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setMinutes(min: number, sec?: number, ms?: number): number` | `setMinutes(min: number, sec?: number, ms?: number): number` | `__date.setMinutes` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setMonth(month: number, date?: number): number` | `setMonth(month: number, date?: number): number` | `__date.setMonth` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setSeconds(sec: number, ms?: number): number` | `setSeconds(sec: number, ms?: number): number` | `__date.setSeconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setTime(time: number): number` | `setTime(time: number): number` | `__date.setTime` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCDate(date: number): number` | `setUTCDate(date: number): number` | `__date.setUTCDate` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCFullYear(year: number, month?: number, date?: number): number` | `setUTCFullYear(year: number, month?: number, date?: number): number` | `__date.setUTCFullYear` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCHours(hours: number, min?: number, sec?: number, ms?: number): number` | `setUTCHours(hours: number, min?: number, sec?: number, ms?: number): number` | `__date.setUTCHours` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCMilliseconds(ms: number): number` | `setUTCMilliseconds(ms: number): number` | `__date.setUTCMilliseconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCMinutes(min: number, sec?: number, ms?: number): number` | `setUTCMinutes(min: number, sec?: number, ms?: number): number` | `__date.setUTCMinutes` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCMonth(month: number, date?: number): number` | `setUTCMonth(month: number, date?: number): number` | `__date.setUTCMonth` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.setUTCSeconds(sec: number, ms?: number): number` | `setUTCSeconds(sec: number, ms?: number): number` | `__date.setUTCSeconds` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toDateString(): string` | `toDateString(): string` | `__date.toDateString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toISOString(): string` | `toISOString(): string` | `__date.toISOString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toJSON(key?: any): string` | `toJSON(key?: any): string` | `__date.toJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toLocaleDateString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleDateString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleDateString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toLocaleTimeString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleTimeString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleTimeString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toString(): string` | `toString(): string` | `__date.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toTemporalInstant(): Temporal.Instant` | `toTemporalInstant(): Temporal.Instant` | `__date.toTemporalInstant` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toTimeString(): string` | `toTimeString(): string` | `__date.toTimeString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.toUTCString(): string` | `toUTCString(): string` | `__date.toUTCString` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `Date.valueOf(): number` | `valueOf(): number` | `__date.valueOf` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |
| `new Date(value: number \| string \| Date): Date` | `new (value: number \| string \| Date): Date` | `__date.new` | ✅ Done | `internal/compiler/testdata/corpus/api/date.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `date` are organized per API under `internal/compiler/testdata/corpus/date/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/date/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
