# Int32Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Int32Array Specification](https://tc39.es/ecma262/#sec-int32array-objects)  
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
| `Int32Array.from(elements: Iterable<number>): Int32Array<ArrayBuffer>` | `from(elements: Iterable<number>): Int32Array<ArrayBuffer>` | `__int32array.from` | ✅ Done | `internal/compiler/testdata/corpus/api/int32array/from/` |
| `Int32Array.from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Int32Array<ArrayBuffer>` | `from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Int32Array<ArrayBuffer>` | `__int32array.from<T>` | ✅ Done | `internal/compiler/testdata/corpus/api/int32array/from/` |
| `Int32Array.): S \| undefined` | `): S \| undefined` | `__int32array.)` | 📋 Planned | - |
| `Int32Array.array: this,` | `array: this,` | `__int32array.array` | 📋 Planned | - |
| `Int32Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__int32array.at` | 📋 Planned | - |
| `Int32Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__int32array.copyWithin` | 📋 Planned | - |
| `Int32Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__int32array.entries` | 📋 Planned | - |
| `Int32Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int32array.every` | 📋 Planned | - |
| `Int32Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__int32array.fill` | 📋 Planned | - |
| `Int32Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int32Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int32Array<ArrayBuffer>` | `__int32array.filter` | 📋 Planned | - |
| `Int32Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__int32array.find` | 📋 Planned | - |
| `Int32Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__int32array.findIndex` | 📋 Planned | - |
| `Int32Array.findLast(` | `findLast(` | `__int32array.findLast` | 📋 Planned | - |
| `Int32Array.findLastIndex(` | `findLastIndex(` | `__int32array.findLastIndex` | 📋 Planned | - |
| `Int32Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__int32array.forEach` | 📋 Planned | - |
| `Int32Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__int32array.includes` | 📋 Planned | - |
| `Int32Array.index: number,` | `index: number,` | `__int32array.index` | 📋 Planned | - |
| `Int32Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__int32array.indexOf` | 📋 Planned | - |
| `Int32Array.join(separator?: string): string` | `join(separator?: string): string` | `__int32array.join` | 📋 Planned | - |
| `Int32Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__int32array.keys` | 📋 Planned | - |
| `Int32Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__int32array.lastIndexOf` | 📋 Planned | - |
| `Int32Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int32Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int32Array<ArrayBuffer>` | `__int32array.map` | 📋 Planned | - |
| `Int32Array.new (elements: Iterable<number>): Int32Array<ArrayBuffer>` | `new (elements: Iterable<number>): Int32Array<ArrayBuffer>` | `__int32array.new` | 📋 Planned | - |
| `Int32Array.of(...items: number[]): Int32Array<ArrayBuffer>` | `of(...items: number[]): Int32Array<ArrayBuffer>` | `__int32array.of` | 📋 Planned | - |
| `Int32Array.predicate: (` | `predicate: (` | `__int32array.predicate` | 📋 Planned | - |
| `Int32Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__int32array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Int32Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__int32array.buffer` | 📋 Planned | - |
| `Int32Array.readonly byteLength: number` | `readonly byteLength: number` | `__int32array.byteLength` | 📋 Planned | - |
| `Int32Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__int32array.byteOffset` | 📋 Planned | - |
| `Int32Array.readonly length: number` | `readonly length: number` | `__int32array.length` | 📋 Planned | - |
| `Int32Array.readonly prototype: Int32Array<ArrayBufferLike>` | `readonly prototype: Int32Array<ArrayBufferLike>` | `__int32array.prototype` | 📋 Planned | - |
| `Int32Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int32array.reduce` | 📋 Planned | - |
| `Int32Array.reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__int32array.reduce<U>` | 📋 Planned | - |
| `Int32Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int32array.reduceRight` | 📋 Planned | - |
| `Int32Array.reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__int32array.reduceRight<U>` | 📋 Planned | - |
| `Int32Array.reverse(): this` | `reverse(): this` | `__int32array.reverse` | 📋 Planned | - |
| `Int32Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__int32array.set` | 📋 Planned | - |
| `Int32Array.slice(start?: number, end?: number): Int32Array<ArrayBuffer>` | `slice(start?: number, end?: number): Int32Array<ArrayBuffer>` | `__int32array.slice` | 📋 Planned | - |
| `Int32Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int32array.some` | 📋 Planned | - |
| `Int32Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__int32array.sort` | 📋 Planned | - |
| `Int32Array.subarray(begin?: number, end?: number): Int32Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Int32Array<TArrayBuffer>` | `__int32array.subarray` | 📋 Planned | - |
| `Int32Array.thisArg?: any,` | `thisArg?: any,` | `__int32array.thisArg?` | 📋 Planned | - |
| `Int32Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__int32array.toLocaleString` | 📋 Planned | - |
| `Int32Array.toReversed(): Int32Array<ArrayBuffer>` | `toReversed(): Int32Array<ArrayBuffer>` | `__int32array.toReversed` | 📋 Planned | - |
| `Int32Array.toSorted(compareFn?: (a: number, b: number) => number): Int32Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Int32Array<ArrayBuffer>` | `__int32array.toSorted` | 📋 Planned | - |
| `Int32Array.toString(): string` | `toString(): string` | `__int32array.toString` | 📋 Planned | - |
| `Int32Array.value: number,` | `value: number,` | `__int32array.value` | 📋 Planned | - |
| `Int32Array.valueOf(): this` | `valueOf(): this` | `__int32array.valueOf` | 📋 Planned | - |
| `Int32Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__int32array.values` | 📋 Planned | - |
| `Int32Array.with(index: number, value: number): Int32Array<ArrayBuffer>` | `with(index: number, value: number): Int32Array<ArrayBuffer>` | `__int32array.with` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `int32array` are organized per API under `internal/compiler/testdata/corpus/int32array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/int32array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
