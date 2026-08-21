# Uint16Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Uint16Array Specification](https://tc39.es/ecma262/#sec-uint16array-objects)  
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
| `Uint16Array.): S \| undefined` | `): S \| undefined` | `__uint16array.)` | 📋 Planned | - |
| `Uint16Array.array: this,` | `array: this,` | `__uint16array.array` | 📋 Planned | - |
| `Uint16Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__uint16array.at` | 📋 Planned | - |
| `Uint16Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__uint16array.copyWithin` | 📋 Planned | - |
| `Uint16Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__uint16array.entries` | 📋 Planned | - |
| `Uint16Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint16array.every` | 📋 Planned | - |
| `Uint16Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__uint16array.fill` | 📋 Planned | - |
| `Uint16Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint16Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint16Array<ArrayBuffer>` | `__uint16array.filter` | 📋 Planned | - |
| `Uint16Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__uint16array.find` | 📋 Planned | - |
| `Uint16Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__uint16array.findIndex` | 📋 Planned | - |
| `Uint16Array.findLast(` | `findLast(` | `__uint16array.findLast` | 📋 Planned | - |
| `Uint16Array.findLastIndex(` | `findLastIndex(` | `__uint16array.findLastIndex` | 📋 Planned | - |
| `Uint16Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__uint16array.forEach` | 📋 Planned | - |
| `Uint16Array.from(elements: Iterable<number>): Uint16Array<ArrayBuffer>` | `from(elements: Iterable<number>): Uint16Array<ArrayBuffer>` | `__uint16array.from` | 📋 Planned | - |
| `Uint16Array.from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Uint16Array<ArrayBuffer>` | `from<T>(elements: Iterable<T>, mapfn?: (v: T, k: number) => number, thisArg?: any): Uint16Array<ArrayBuffer>` | `__uint16array.from<T>` | 📋 Planned | - |
| `Uint16Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__uint16array.includes` | 📋 Planned | - |
| `Uint16Array.index: number,` | `index: number,` | `__uint16array.index` | 📋 Planned | - |
| `Uint16Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__uint16array.indexOf` | 📋 Planned | - |
| `Uint16Array.join(separator?: string): string` | `join(separator?: string): string` | `__uint16array.join` | 📋 Planned | - |
| `Uint16Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__uint16array.keys` | 📋 Planned | - |
| `Uint16Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__uint16array.lastIndexOf` | 📋 Planned | - |
| `Uint16Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint16Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint16Array<ArrayBuffer>` | `__uint16array.map` | 📋 Planned | - |
| `Uint16Array.new (elements: Iterable<number>): Uint16Array<ArrayBuffer>` | `new (elements: Iterable<number>): Uint16Array<ArrayBuffer>` | `__uint16array.new` | 📋 Planned | - |
| `Uint16Array.of(...items: number[]): Uint16Array<ArrayBuffer>` | `of(...items: number[]): Uint16Array<ArrayBuffer>` | `__uint16array.of` | 📋 Planned | - |
| `Uint16Array.predicate: (` | `predicate: (` | `__uint16array.predicate` | 📋 Planned | - |
| `Uint16Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__uint16array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Uint16Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__uint16array.buffer` | 📋 Planned | - |
| `Uint16Array.readonly byteLength: number` | `readonly byteLength: number` | `__uint16array.byteLength` | 📋 Planned | - |
| `Uint16Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__uint16array.byteOffset` | 📋 Planned | - |
| `Uint16Array.readonly length: number` | `readonly length: number` | `__uint16array.length` | 📋 Planned | - |
| `Uint16Array.readonly prototype: Uint16Array<ArrayBufferLike>` | `readonly prototype: Uint16Array<ArrayBufferLike>` | `__uint16array.prototype` | 📋 Planned | - |
| `Uint16Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint16array.reduce` | 📋 Planned | - |
| `Uint16Array.reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__uint16array.reduce<U>` | 📋 Planned | - |
| `Uint16Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint16array.reduceRight` | 📋 Planned | - |
| `Uint16Array.reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: number, currentIndex: number, array: this) => U, initialValue: U): U` | `__uint16array.reduceRight<U>` | 📋 Planned | - |
| `Uint16Array.reverse(): this` | `reverse(): this` | `__uint16array.reverse` | 📋 Planned | - |
| `Uint16Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__uint16array.set` | 📋 Planned | - |
| `Uint16Array.slice(start?: number, end?: number): Uint16Array<ArrayBuffer>` | `slice(start?: number, end?: number): Uint16Array<ArrayBuffer>` | `__uint16array.slice` | 📋 Planned | - |
| `Uint16Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint16array.some` | 📋 Planned | - |
| `Uint16Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__uint16array.sort` | 📋 Planned | - |
| `Uint16Array.subarray(begin?: number, end?: number): Uint16Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Uint16Array<TArrayBuffer>` | `__uint16array.subarray` | 📋 Planned | - |
| `Uint16Array.thisArg?: any,` | `thisArg?: any,` | `__uint16array.thisArg?` | 📋 Planned | - |
| `Uint16Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__uint16array.toLocaleString` | 📋 Planned | - |
| `Uint16Array.toReversed(): Uint16Array<ArrayBuffer>` | `toReversed(): Uint16Array<ArrayBuffer>` | `__uint16array.toReversed` | 📋 Planned | - |
| `Uint16Array.toSorted(compareFn?: (a: number, b: number) => number): Uint16Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Uint16Array<ArrayBuffer>` | `__uint16array.toSorted` | 📋 Planned | - |
| `Uint16Array.toString(): string` | `toString(): string` | `__uint16array.toString` | 📋 Planned | - |
| `Uint16Array.value: number,` | `value: number,` | `__uint16array.value` | 📋 Planned | - |
| `Uint16Array.valueOf(): this` | `valueOf(): this` | `__uint16array.valueOf` | 📋 Planned | - |
| `Uint16Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__uint16array.values` | 📋 Planned | - |
| `Uint16Array.with(index: number, value: number): Uint16Array<ArrayBuffer>` | `with(index: number, value: number): Uint16Array<ArrayBuffer>` | `__uint16array.with` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `uint16array` are organized per API under `internal/compiler/testdata/corpus/uint16array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/uint16array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
