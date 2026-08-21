# Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Array Specification](https://tc39.es/ecma262/#sec-array-objects)  
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
| `Array.concat(...items: ConcatArray<T>[]): T[]` | `concat(...items: ConcatArray<T>[]): T[]` | `__array.concat` | ✅ Done | `internal/compiler/testdata/corpus/api/array/concat/` |
| `Array.filter(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): T[]` | `filter(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): T[]` | `__array.filter` | ✅ Done | `internal/compiler/testdata/corpus/api/array/filter/` |
| `Array.forEach(callbackfn: (value: T, index: number, array: T[]) => void, thisArg?: any): void` | `forEach(callbackfn: (value: T, index: number, array: T[]) => void, thisArg?: any): void` | `__array.forEach` | ✅ Done | `internal/compiler/testdata/corpus/api/array/forEach/` |
| `Array.includes(searchElement: T, fromIndex?: number): boolean` | `includes(searchElement: T, fromIndex?: number): boolean` | `__array.includes` | ✅ Done | `internal/compiler/testdata/corpus/api/array/includes/` |
| `Array.indexOf(searchElement: T, fromIndex?: number): number` | `indexOf(searchElement: T, fromIndex?: number): number` | `__array.indexOf` | ✅ Done | `internal/compiler/testdata/corpus/api/array/indexOf/` |
| `Array.join(separator?: string): string` | `join(separator?: string): string` | `__array.join` | ✅ Done | `internal/compiler/testdata/corpus/api/array/join/` |
| `Array.map<U>(callbackfn: (value: T, index: number, array: T[]) => U, thisArg?: any): U[]` | `map<U>(callbackfn: (value: T, index: number, array: T[]) => U, thisArg?: any): U[]` | `__array.map<U>` | ✅ Done | `internal/compiler/testdata/corpus/api/array/map/` |
| `Array.pop(): T \| undefined` | `pop(): T \| undefined` | `__array.pop` | ✅ Done | `internal/compiler/testdata/corpus/api/array/pop/` |
| `Array.push(...items: T[]): number` | `push(...items: T[]): number` | `__array.push` | ✅ Done | `internal/compiler/testdata/corpus/api/array/push/` |
| `Array.reduce(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: T[]) => T): T` | `reduce(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: T[]) => T): T` | `__array.reduce` | ✅ Done | `internal/compiler/testdata/corpus/api/array/reduce/` |
| `Array.reduce<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: T[]) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: T[]) => U, initialValue: U): U` | `__array.reduce<U>` | ✅ Done | `internal/compiler/testdata/corpus/api/array/reduce/` |
| `Array.shift(): T \| undefined` | `shift(): T \| undefined` | `__array.shift` | ✅ Done | `internal/compiler/testdata/corpus/api/array/shift/` |
| `Array.slice(start?: number, end?: number): T[]` | `slice(start?: number, end?: number): T[]` | `__array.slice` | ✅ Done | `internal/compiler/testdata/corpus/api/array/slice/` |
| `Array.splice(start: number, deleteCount?: number): T[]` | `splice(start: number, deleteCount?: number): T[]` | `__array.splice` | ✅ Done | `internal/compiler/testdata/corpus/api/array/splice/` |
| `Array.unshift(...items: T[]): number` | `unshift(...items: T[]): number` | `__array.unshift` | ✅ Done | `internal/compiler/testdata/corpus/api/array/unshift/` |
| `Array.): U[]` | `): U[]` | `__array.)` | 📋 Planned | - |
| `Array.<T>(arrayLength: number): T[]` | `<T>(arrayLength: number): T[]` | `__array.<T>` | 📋 Planned | - |
| `Array.at(index: number): T \| undefined` | `at(index: number): T \| undefined` | `__array.at` | 📋 Planned | - |
| `Array.callback: (this: This, value: T, index: number, array: T[]) => U \| ReadonlyArray<U>,` | `callback: (this: This, value: T, index: number, array: T[]) => U \| ReadonlyArray<U>,` | `__array.callback` | 📋 Planned | - |
| `Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__array.copyWithin` | 📋 Planned | - |
| `Array.depth?: D,` | `depth?: D,` | `__array.depth?` | 📋 Planned | - |
| `Array.entries(): ArrayIterator<[number, T]>` | `entries(): ArrayIterator<[number, T]>` | `__array.entries` | 📋 Planned | - |
| `Array.every(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): boolean` | `every(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): boolean` | `__array.every` | 📋 Planned | - |
| `Array.fill(value: T, start?: number, end?: number): this` | `fill(value: T, start?: number, end?: number): this` | `__array.fill` | 📋 Planned | - |
| `Array.find(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): T \| undefined` | `find(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): T \| undefined` | `__array.find` | 📋 Planned | - |
| `Array.findIndex(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): number` | `findIndex(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): number` | `__array.findIndex` | 📋 Planned | - |
| `Array.findLast(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): T \| undefined` | `findLast(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): T \| undefined` | `__array.findLast` | 📋 Planned | - |
| `Array.findLastIndex(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): number` | `findLastIndex(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): number` | `__array.findLastIndex` | 📋 Planned | - |
| `Array.from<T>(arrayLike: ArrayLike<T>): T[]` | `from<T>(arrayLike: ArrayLike<T>): T[]` | `__array.from<T>` | 📋 Planned | - |
| `Array.fromAsync<T>(iterableOrArrayLike: AsyncIterable<T> \| Iterable<T \| PromiseLike<T>> \| ArrayLike<T \| PromiseLike<T>>): Promise<T[]>` | `fromAsync<T>(iterableOrArrayLike: AsyncIterable<T> \| Iterable<T \| PromiseLike<T>> \| ArrayLike<T \| PromiseLike<T>>): Promise<T[]>` | `__array.fromAsync<T>` | 📋 Planned | - |
| `Array.isArray(arg: any): arg is any[]` | `isArray(arg: any): arg is any[]` | `__array.isArray` | 📋 Planned | - |
| `Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__array.keys` | 📋 Planned | - |
| `Array.lastIndexOf(searchElement: T, fromIndex?: number): number` | `lastIndexOf(searchElement: T, fromIndex?: number): number` | `__array.lastIndexOf` | 📋 Planned | - |
| `Array.length: number` | `length: number` | `__array.length` | 📋 Planned | - |
| `Array.new (arrayLength?: number): any[]` | `new (arrayLength?: number): any[]` | `__array.new` | 📋 Planned | - |
| `Array.of<T>(...items: T[]): T[]` | `of<T>(...items: T[]): T[]` | `__array.of<T>` | 📋 Planned | - |
| `Array.readonly prototype: any[]` | `readonly prototype: any[]` | `__array.prototype` | 📋 Planned | - |
| `Array.reduceRight(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: T[]) => T): T` | `reduceRight(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: T[]) => T): T` | `__array.reduceRight` | 📋 Planned | - |
| `Array.reduceRight<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: T[]) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: T[]) => U, initialValue: U): U` | `__array.reduceRight<U>` | 📋 Planned | - |
| `Array.reverse(): T[]` | `reverse(): T[]` | `__array.reverse` | 📋 Planned | - |
| `Array.some(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): boolean` | `some(predicate: (value: T, index: number, array: T[]) => unknown, thisArg?: any): boolean` | `__array.some` | 📋 Planned | - |
| `Array.sort(compareFn?: (a: T, b: T) => number): this` | `sort(compareFn?: (a: T, b: T) => number): this` | `__array.sort` | 📋 Planned | - |
| `Array.this: A,` | `this: A,` | `__array.this` | 📋 Planned | - |
| `Array.thisArg?: This,` | `thisArg?: This,` | `__array.thisArg?` | 📋 Planned | - |
| `Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions & Intl.DateTimeFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions & Intl.DateTimeFormatOptions): string` | `__array.toLocaleString` | 📋 Planned | - |
| `Array.toReversed(): T[]` | `toReversed(): T[]` | `__array.toReversed` | 📋 Planned | - |
| `Array.toSorted(compareFn?: (a: T, b: T) => number): T[]` | `toSorted(compareFn?: (a: T, b: T) => number): T[]` | `__array.toSorted` | 📋 Planned | - |
| `Array.toSpliced(start: number, deleteCount: number, ...items: T[]): T[]` | `toSpliced(start: number, deleteCount: number, ...items: T[]): T[]` | `__array.toSpliced` | 📋 Planned | - |
| `Array.toString(): string` | `toString(): string` | `__array.toString` | 📋 Planned | - |
| `Array.values(): ArrayIterator<T>` | `values(): ArrayIterator<T>` | `__array.values` | 📋 Planned | - |
| `Array.with(index: number, value: T): T[]` | `with(index: number, value: T): T[]` | `__array.with` | 📋 Planned | - |
| `Array.}` | `}` | `__array.}` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `array` are organized per API under `internal/compiler/testdata/corpus/array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
