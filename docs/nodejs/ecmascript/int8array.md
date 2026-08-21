# Int8Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Int8Array Specification](https://tc39.es/ecma262/#sec-int8array-objects)  
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
| `Int8Array.): S \| undefined` | `): S \| undefined` | `__int8array.)` | 📋 Planned | - |
| `Int8Array.array: this,` | `array: this,` | `__int8array.array` | 📋 Planned | - |
| `Int8Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__int8array.at` | 📋 Planned | - |
| `Int8Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__int8array.copyWithin` | 📋 Planned | - |
| `Int8Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__int8array.entries` | 📋 Planned | - |
| `Int8Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int8array.every` | 📋 Planned | - |
| `Int8Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__int8array.fill` | 📋 Planned | - |
| `Int8Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int8Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int8Array<ArrayBuffer>` | `__int8array.filter` | 📋 Planned | - |
| `Int8Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__int8array.find` | 📋 Planned | - |
| `Int8Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__int8array.findIndex` | 📋 Planned | - |
| `Int8Array.findLast(` | `findLast(` | `__int8array.findLast` | 📋 Planned | - |
| `Int8Array.findLastIndex(` | `findLastIndex(` | `__int8array.findLastIndex` | 📋 Planned | - |
| `Int8Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__int8array.forEach` | 📋 Planned | - |
| `Int8Array.from(elements: Iterable<number>): Int8Array<ArrayBuffer>` | `from(elements: Iterable<number>): Int8Array<ArrayBuffer>` | `__int8array.from` | 📋 Planned | - |
| `Int8Array.from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Int8Array<ArrayBuffer>` | `from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Int8Array<ArrayBuffer>` | `__int8array.from<T>` | 📋 Planned | - |
| `Int8Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__int8array.includes` | 📋 Planned | - |
| `Int8Array.index: number,` | `index: number,` | `__int8array.index` | 📋 Planned | - |
| `Int8Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__int8array.indexOf` | 📋 Planned | - |
| `Int8Array.join(separator?: string): string` | `join(separator?: string): string` | `__int8array.join` | 📋 Planned | - |
| `Int8Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__int8array.keys` | 📋 Planned | - |
| `Int8Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__int8array.lastIndexOf` | 📋 Planned | - |
| `Int8Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int8Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int8Array<ArrayBuffer>` | `__int8array.map` | 📋 Planned | - |
| `Int8Array.new (elements: Iterable<number>): Int8Array<ArrayBuffer>` | `new (elements: Iterable<number>): Int8Array<ArrayBuffer>` | `__int8array.new` | 📋 Planned | - |
| `Int8Array.of(...items: number[]): Int8Array<ArrayBuffer>` | `of(...items: number[]): Int8Array<ArrayBuffer>` | `__int8array.of` | 📋 Planned | - |
| `Int8Array.predicate: (` | `predicate: (` | `__int8array.predicate` | 📋 Planned | - |
| `Int8Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__int8array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Int8Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__int8array.buffer` | 📋 Planned | - |
| `Int8Array.readonly byteLength: number` | `readonly byteLength: number` | `__int8array.byteLength` | 📋 Planned | - |
| `Int8Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__int8array.byteOffset` | 📋 Planned | - |
| `Int8Array.readonly length: number` | `readonly length: number` | `__int8array.length` | 📋 Planned | - |
| `Int8Array.readonly prototype: Int8Array<ArrayBufferLike>` | `readonly prototype: Int8Array<ArrayBufferLike>` | `__int8array.prototype` | 📋 Planned | - |
| `Int8Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int8array.reduce` | 📋 Planned | - |
| `Int8Array.reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__int8array.reduce<U>` | 📋 Planned | - |
| `Int8Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int8array.reduceRight` | 📋 Planned | - |
| `Int8Array.reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__int8array.reduceRight<U>` | 📋 Planned | - |
| `Int8Array.reverse(): this` | `reverse(): this` | `__int8array.reverse` | 📋 Planned | - |
| `Int8Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__int8array.set` | 📋 Planned | - |
| `Int8Array.slice(start?: number, end?: number): Int8Array<ArrayBuffer>` | `slice(start?: number, end?: number): Int8Array<ArrayBuffer>` | `__int8array.slice` | 📋 Planned | - |
| `Int8Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int8array.some` | 📋 Planned | - |
| `Int8Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__int8array.sort` | 📋 Planned | - |
| `Int8Array.subarray(begin?: number, end?: number): Int8Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Int8Array<TArrayBuffer>` | `__int8array.subarray` | 📋 Planned | - |
| `Int8Array.thisArg?: any,` | `thisArg?: any,` | `__int8array.thisArg?` | 📋 Planned | - |
| `Int8Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__int8array.toLocaleString` | 📋 Planned | - |
| `Int8Array.toReversed(): Int8Array<ArrayBuffer>` | `toReversed(): Int8Array<ArrayBuffer>` | `__int8array.toReversed` | 📋 Planned | - |
| `Int8Array.toSorted(compareFn?: (a: number, b: number) => number): Int8Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Int8Array<ArrayBuffer>` | `__int8array.toSorted` | 📋 Planned | - |
| `Int8Array.toString(): string` | `toString(): string` | `__int8array.toString` | 📋 Planned | - |
| `Int8Array.value: number,` | `value: number,` | `__int8array.value` | 📋 Planned | - |
| `Int8Array.valueOf(): this` | `valueOf(): this` | `__int8array.valueOf` | 📋 Planned | - |
| `Int8Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__int8array.values` | 📋 Planned | - |
| `Int8Array.with(index: number, value: number): Int8Array<ArrayBuffer>` | `with(index: number, value: number): Int8Array<ArrayBuffer>` | `__int8array.with` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `int8array` are organized per API under `internal/compiler/testdata/corpus/int8array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/int8array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
