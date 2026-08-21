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
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `Date.getTime(): number` | `getTime(): number` | `__date.getTime` | ✅ Done | `internal/compiler/testdata/corpus/api/date/getTime/` |
| `Date.now(): number` | `now(): number` | `__date.now` | ✅ Done | `internal/compiler/testdata/corpus/api/date/now/` |
| `Date.parse(s: string): number` | `parse(s: string): number` | `__date.parse` | ✅ Done | `internal/compiler/testdata/corpus/api/date/parse/` |
| `Date.toISOString(): string` | `toISOString(): string` | `__date.toISOString` | ✅ Done | `internal/compiler/testdata/corpus/api/date/toISOString/` |
| `Date.toString(): string` | `toString(): string` | `__date.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/date/toString/` |
| `Date.UTC(year: number, monthIndex?: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): number` | `UTC(year: number, monthIndex?: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): number` | `__date.UTC` | 📋 Planned | - |
| `Date.getDate(): number` | `getDate(): number` | `__date.getDate` | 📋 Planned | - |
| `Date.getDay(): number` | `getDay(): number` | `__date.getDay` | 📋 Planned | - |
| `Date.getFullYear(): number` | `getFullYear(): number` | `__date.getFullYear` | 📋 Planned | - |
| `Date.getHours(): number` | `getHours(): number` | `__date.getHours` | 📋 Planned | - |
| `Date.getMilliseconds(): number` | `getMilliseconds(): number` | `__date.getMilliseconds` | 📋 Planned | - |
| `Date.getMinutes(): number` | `getMinutes(): number` | `__date.getMinutes` | 📋 Planned | - |
| `Date.getMonth(): number` | `getMonth(): number` | `__date.getMonth` | 📋 Planned | - |
| `Date.getSeconds(): number` | `getSeconds(): number` | `__date.getSeconds` | 📋 Planned | - |
| `Date.getTimezoneOffset(): number` | `getTimezoneOffset(): number` | `__date.getTimezoneOffset` | 📋 Planned | - |
| `Date.getUTCDate(): number` | `getUTCDate(): number` | `__date.getUTCDate` | 📋 Planned | - |
| `Date.getUTCDay(): number` | `getUTCDay(): number` | `__date.getUTCDay` | 📋 Planned | - |
| `Date.getUTCFullYear(): number` | `getUTCFullYear(): number` | `__date.getUTCFullYear` | 📋 Planned | - |
| `Date.getUTCHours(): number` | `getUTCHours(): number` | `__date.getUTCHours` | 📋 Planned | - |
| `Date.getUTCMilliseconds(): number` | `getUTCMilliseconds(): number` | `__date.getUTCMilliseconds` | 📋 Planned | - |
| `Date.getUTCMinutes(): number` | `getUTCMinutes(): number` | `__date.getUTCMinutes` | 📋 Planned | - |
| `Date.getUTCMonth(): number` | `getUTCMonth(): number` | `__date.getUTCMonth` | 📋 Planned | - |
| `Date.getUTCSeconds(): number` | `getUTCSeconds(): number` | `__date.getUTCSeconds` | 📋 Planned | - |
| `Date.new (value: number \| string \| Date): Date` | `new (value: number \| string \| Date): Date` | `__date.new` | 📋 Planned | - |
| `Date.readonly prototype: Date` | `readonly prototype: Date` | `__date.prototype` | 📋 Planned | - |
| `Date.setDate(date: number): number` | `setDate(date: number): number` | `__date.setDate` | 📋 Planned | - |
| `Date.setFullYear(year: number, month?: number, date?: number): number` | `setFullYear(year: number, month?: number, date?: number): number` | `__date.setFullYear` | 📋 Planned | - |
| `Date.setHours(hours: number, min?: number, sec?: number, ms?: number): number` | `setHours(hours: number, min?: number, sec?: number, ms?: number): number` | `__date.setHours` | 📋 Planned | - |
| `Date.setMilliseconds(ms: number): number` | `setMilliseconds(ms: number): number` | `__date.setMilliseconds` | 📋 Planned | - |
| `Date.setMinutes(min: number, sec?: number, ms?: number): number` | `setMinutes(min: number, sec?: number, ms?: number): number` | `__date.setMinutes` | 📋 Planned | - |
| `Date.setMonth(month: number, date?: number): number` | `setMonth(month: number, date?: number): number` | `__date.setMonth` | 📋 Planned | - |
| `Date.setSeconds(sec: number, ms?: number): number` | `setSeconds(sec: number, ms?: number): number` | `__date.setSeconds` | 📋 Planned | - |
| `Date.setTime(time: number): number` | `setTime(time: number): number` | `__date.setTime` | 📋 Planned | - |
| `Date.setUTCDate(date: number): number` | `setUTCDate(date: number): number` | `__date.setUTCDate` | 📋 Planned | - |
| `Date.setUTCFullYear(year: number, month?: number, date?: number): number` | `setUTCFullYear(year: number, month?: number, date?: number): number` | `__date.setUTCFullYear` | 📋 Planned | - |
| `Date.setUTCHours(hours: number, min?: number, sec?: number, ms?: number): number` | `setUTCHours(hours: number, min?: number, sec?: number, ms?: number): number` | `__date.setUTCHours` | 📋 Planned | - |
| `Date.setUTCMilliseconds(ms: number): number` | `setUTCMilliseconds(ms: number): number` | `__date.setUTCMilliseconds` | 📋 Planned | - |
| `Date.setUTCMinutes(min: number, sec?: number, ms?: number): number` | `setUTCMinutes(min: number, sec?: number, ms?: number): number` | `__date.setUTCMinutes` | 📋 Planned | - |
| `Date.setUTCMonth(month: number, date?: number): number` | `setUTCMonth(month: number, date?: number): number` | `__date.setUTCMonth` | 📋 Planned | - |
| `Date.setUTCSeconds(sec: number, ms?: number): number` | `setUTCSeconds(sec: number, ms?: number): number` | `__date.setUTCSeconds` | 📋 Planned | - |
| `Date.toDateString(): string` | `toDateString(): string` | `__date.toDateString` | 📋 Planned | - |
| `Date.toJSON(key?: any): string` | `toJSON(key?: any): string` | `__date.toJSON` | 📋 Planned | - |
| `Date.toLocaleDateString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleDateString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleDateString` | 📋 Planned | - |
| `Date.toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleString` | 📋 Planned | - |
| `Date.toLocaleTimeString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `toLocaleTimeString(locales?: Intl.LocalesArgument, options?: Intl.DateTimeFormatOptions): string` | `__date.toLocaleTimeString` | 📋 Planned | - |
| `Date.toTemporalInstant(): Temporal.Instant` | `toTemporalInstant(): Temporal.Instant` | `__date.toTemporalInstant` | 📋 Planned | - |
| `Date.toTimeString(): string` | `toTimeString(): string` | `__date.toTimeString` | 📋 Planned | - |
| `Date.toUTCString(): string` | `toUTCString(): string` | `__date.toUTCString` | 📋 Planned | - |
| `Date.valueOf(): number` | `valueOf(): number` | `__date.valueOf` | 📋 Planned | - |

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
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/date/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
