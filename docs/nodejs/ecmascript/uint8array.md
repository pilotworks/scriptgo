# Uint8Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Uint8Array Specification](https://tc39.es/ecma262/#sec-uint8array-objects)  
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
| `Uint8Array.from(elements: Iterable<number>): Uint8Array<ArrayBuffer>` | `from(elements: Iterable<number>): Uint8Array<ArrayBuffer>` | `__uint8array.from` | ✅ Done | `internal/compiler/testdata/corpus/api/uint8array/from/` |
| `Uint8Array.from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Uint8Array<ArrayBuffer>` | `from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Uint8Array<ArrayBuffer>` | `__uint8array.from<T>` | ✅ Done | `internal/compiler/testdata/corpus/api/uint8array/from/` |
| `Uint8Array.): S \| undefined` | `): S \| undefined` | `__uint8array.)` | 📋 Planned | - |
| `Uint8Array.alphabet?: "base64" \| "base64url" \| undefined` | `alphabet?: "base64" \| "base64url" \| undefined` | `__uint8array.alphabet?` | 📋 Planned | - |
| `Uint8Array.array: this,` | `array: this,` | `__uint8array.array` | 📋 Planned | - |
| `Uint8Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__uint8array.at` | 📋 Planned | - |
| `Uint8Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__uint8array.copyWithin` | 📋 Planned | - |
| `Uint8Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__uint8array.entries` | 📋 Planned | - |
| `Uint8Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint8array.every` | 📋 Planned | - |
| `Uint8Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__uint8array.fill` | 📋 Planned | - |
| `Uint8Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint8Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint8Array<ArrayBuffer>` | `__uint8array.filter` | 📋 Planned | - |
| `Uint8Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__uint8array.find` | 📋 Planned | - |
| `Uint8Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__uint8array.findIndex` | 📋 Planned | - |
| `Uint8Array.findLast(` | `findLast(` | `__uint8array.findLast` | 📋 Planned | - |
| `Uint8Array.findLastIndex(` | `findLastIndex(` | `__uint8array.findLastIndex` | 📋 Planned | - |
| `Uint8Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__uint8array.forEach` | 📋 Planned | - |
| `Uint8Array.fromBase64(` | `fromBase64(` | `__uint8array.fromBase64` | 📋 Planned | - |
| `Uint8Array.fromHex(` | `fromHex(` | `__uint8array.fromHex` | 📋 Planned | - |
| `Uint8Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__uint8array.includes` | 📋 Planned | - |
| `Uint8Array.index: number,` | `index: number,` | `__uint8array.index` | 📋 Planned | - |
| `Uint8Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__uint8array.indexOf` | 📋 Planned | - |
| `Uint8Array.join(separator?: string): string` | `join(separator?: string): string` | `__uint8array.join` | 📋 Planned | - |
| `Uint8Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__uint8array.keys` | 📋 Planned | - |
| `Uint8Array.lastChunkHandling?: "loose" \| "strict" \| "stop-before-partial" \| undefined` | `lastChunkHandling?: "loose" \| "strict" \| "stop-before-partial" \| undefined` | `__uint8array.lastChunkHandling?` | 📋 Planned | - |
| `Uint8Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__uint8array.lastIndexOf` | 📋 Planned | - |
| `Uint8Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint8Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint8Array<ArrayBuffer>` | `__uint8array.map` | 📋 Planned | - |
| `Uint8Array.new (elements: Iterable<number>): Uint8Array<ArrayBuffer>` | `new (elements: Iterable<number>): Uint8Array<ArrayBuffer>` | `__uint8array.new` | 📋 Planned | - |
| `Uint8Array.of(...items: number[]): Uint8Array<ArrayBuffer>` | `of(...items: number[]): Uint8Array<ArrayBuffer>` | `__uint8array.of` | 📋 Planned | - |
| `Uint8Array.omitPadding?: boolean \| undefined` | `omitPadding?: boolean \| undefined` | `__uint8array.omitPadding?` | 📋 Planned | - |
| `Uint8Array.options?: {` | `options?: {` | `__uint8array.options?` | 📋 Planned | - |
| `Uint8Array.predicate: (` | `predicate: (` | `__uint8array.predicate` | 📋 Planned | - |
| `Uint8Array.read: number` | `read: number` | `__uint8array.read` | 📋 Planned | - |
| `Uint8Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__uint8array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Uint8Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__uint8array.buffer` | 📋 Planned | - |
| `Uint8Array.readonly byteLength: number` | `readonly byteLength: number` | `__uint8array.byteLength` | 📋 Planned | - |
| `Uint8Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__uint8array.byteOffset` | 📋 Planned | - |
| `Uint8Array.readonly length: number` | `readonly length: number` | `__uint8array.length` | 📋 Planned | - |
| `Uint8Array.readonly prototype: Uint8Array<ArrayBufferLike>` | `readonly prototype: Uint8Array<ArrayBufferLike>` | `__uint8array.prototype` | 📋 Planned | - |
| `Uint8Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint8array.reduce` | 📋 Planned | - |
| `Uint8Array.reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__uint8array.reduce<U>` | 📋 Planned | - |
| `Uint8Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint8array.reduceRight` | 📋 Planned | - |
| `Uint8Array.reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__uint8array.reduceRight<U>` | 📋 Planned | - |
| `Uint8Array.reverse(): this` | `reverse(): this` | `__uint8array.reverse` | 📋 Planned | - |
| `Uint8Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__uint8array.set` | 📋 Planned | - |
| `Uint8Array.setFromBase64(` | `setFromBase64(` | `__uint8array.setFromBase64` | 📋 Planned | - |
| `Uint8Array.setFromHex(string: string): {` | `setFromHex(string: string): {` | `__uint8array.setFromHex` | 📋 Planned | - |
| `Uint8Array.slice(start?: number, end?: number): Uint8Array<ArrayBuffer>` | `slice(start?: number, end?: number): Uint8Array<ArrayBuffer>` | `__uint8array.slice` | 📋 Planned | - |
| `Uint8Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint8array.some` | 📋 Planned | - |
| `Uint8Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__uint8array.sort` | 📋 Planned | - |
| `Uint8Array.string: string,` | `string: string,` | `__uint8array.string` | 📋 Planned | - |
| `Uint8Array.subarray(begin?: number, end?: number): Uint8Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Uint8Array<TArrayBuffer>` | `__uint8array.subarray` | 📋 Planned | - |
| `Uint8Array.thisArg?: any,` | `thisArg?: any,` | `__uint8array.thisArg?` | 📋 Planned | - |
| `Uint8Array.toBase64(` | `toBase64(` | `__uint8array.toBase64` | 📋 Planned | - |
| `Uint8Array.toHex(): string` | `toHex(): string` | `__uint8array.toHex` | 📋 Planned | - |
| `Uint8Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__uint8array.toLocaleString` | 📋 Planned | - |
| `Uint8Array.toReversed(): Uint8Array<ArrayBuffer>` | `toReversed(): Uint8Array<ArrayBuffer>` | `__uint8array.toReversed` | 📋 Planned | - |
| `Uint8Array.toSorted(compareFn?: (a: number, b: number) => number): Uint8Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Uint8Array<ArrayBuffer>` | `__uint8array.toSorted` | 📋 Planned | - |
| `Uint8Array.toString(): string` | `toString(): string` | `__uint8array.toString` | 📋 Planned | - |
| `Uint8Array.value: number,` | `value: number,` | `__uint8array.value` | 📋 Planned | - |
| `Uint8Array.valueOf(): this` | `valueOf(): this` | `__uint8array.valueOf` | 📋 Planned | - |
| `Uint8Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__uint8array.values` | 📋 Planned | - |
| `Uint8Array.with(index: number, value: number): Uint8Array<ArrayBuffer>` | `with(index: number, value: number): Uint8Array<ArrayBuffer>` | `__uint8array.with` | 📋 Planned | - |
| `Uint8Array.written: number` | `written: number` | `__uint8array.written` | 📋 Planned | - |
| `Uint8Array.}` | `}` | `__uint8array.}` | 📋 Planned | - |
| `Uint8Array.},` | `},` | `__uint8array.},` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `uint8array` are organized per API under `internal/compiler/testdata/corpus/uint8array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/uint8array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
