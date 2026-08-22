# Int16Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Int16Array Specification](https://tc39.es/ecma262/#sec-int16array-objects)  
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
| `Int16Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__int16array.at` | 📋 Planned | - |
| `Int16Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__int16array.copyWithin` | 📋 Planned | - |
| `Int16Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__int16array.entries` | 📋 Planned | - |
| `Int16Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int16array.every` | 📋 Planned | - |
| `Int16Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__int16array.fill` | 📋 Planned | - |
| `Int16Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int16Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Int16Array<ArrayBuffer>` | `__int16array.filter` | 📋 Planned | - |
| `Int16Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__int16array.find` | 📋 Planned | - |
| `Int16Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__int16array.findIndex` | 📋 Planned | - |
| `Int16Array.findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__int16array.findLast` | 📋 Planned | - |
| `Int16Array.findLastIndex( predicate: (value: number, index: number, array: this) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: (value: number, index: number, array: this) => unknown, thisArg?: any, ): number` | `__int16array.findLastIndex` | 📋 Planned | - |
| `Int16Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__int16array.forEach` | 📋 Planned | - |
| `Int16Array.from(elements: Iterable<number>): Int16Array<ArrayBuffer>` | `from(elements: Iterable<number>): Int16Array<ArrayBuffer>` | `__int16array.from` | 📋 Planned | - |
| `Int16Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__int16array.includes` | 📋 Planned | - |
| `Int16Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__int16array.indexOf` | 📋 Planned | - |
| `Int16Array.join(separator?: string): string` | `join(separator?: string): string` | `__int16array.join` | 📋 Planned | - |
| `Int16Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__int16array.keys` | 📋 Planned | - |
| `Int16Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__int16array.lastIndexOf` | 📋 Planned | - |
| `Int16Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int16Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Int16Array<ArrayBuffer>` | `__int16array.map` | 📋 Planned | - |
| `Int16Array.new <TArrayBuffer extends ArrayBufferLike = ArrayBuffer>(buffer: TArrayBuffer, byteOffset?: number, length?: number): Int16Array<TArrayBuffer>` | `new <TArrayBuffer extends ArrayBufferLike = ArrayBuffer>(buffer: TArrayBuffer, byteOffset?: number, length?: number): Int16Array<TArrayBuffer>` | `__int16array.new` | 📋 Planned | - |
| `Int16Array.of(...items: number[]): Int16Array<ArrayBuffer>` | `of(...items: number[]): Int16Array<ArrayBuffer>` | `__int16array.of` | 📋 Planned | - |
| `Int16Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__int16array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Int16Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__int16array.buffer` | 📋 Planned | - |
| `Int16Array.readonly byteLength: number` | `readonly byteLength: number` | `__int16array.byteLength` | 📋 Planned | - |
| `Int16Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__int16array.byteOffset` | 📋 Planned | - |
| `Int16Array.readonly length: number` | `readonly length: number` | `__int16array.length` | 📋 Planned | - |
| `Int16Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int16array.reduce` | 📋 Planned | - |
| `Int16Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__int16array.reduceRight` | 📋 Planned | - |
| `Int16Array.reverse(): this` | `reverse(): this` | `__int16array.reverse` | 📋 Planned | - |
| `Int16Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__int16array.set` | 📋 Planned | - |
| `Int16Array.slice(start?: number, end?: number): Int16Array<ArrayBuffer>` | `slice(start?: number, end?: number): Int16Array<ArrayBuffer>` | `__int16array.slice` | 📋 Planned | - |
| `Int16Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__int16array.some` | 📋 Planned | - |
| `Int16Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__int16array.sort` | 📋 Planned | - |
| `Int16Array.subarray(begin?: number, end?: number): Int16Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Int16Array<TArrayBuffer>` | `__int16array.subarray` | 📋 Planned | - |
| `Int16Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__int16array.toLocaleString` | 📋 Planned | - |
| `Int16Array.toReversed(): Int16Array<ArrayBuffer>` | `toReversed(): Int16Array<ArrayBuffer>` | `__int16array.toReversed` | 📋 Planned | - |
| `Int16Array.toSorted(compareFn?: (a: number, b: number) => number): Int16Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Int16Array<ArrayBuffer>` | `__int16array.toSorted` | 📋 Planned | - |
| `Int16Array.toString(): string` | `toString(): string` | `__int16array.toString` | 📋 Planned | - |
| `Int16Array.valueOf(): this` | `valueOf(): this` | `__int16array.valueOf` | 📋 Planned | - |
| `Int16Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__int16array.values` | 📋 Planned | - |
| `Int16Array.with(index: number, value: number): Int16Array<ArrayBuffer>` | `with(index: number, value: number): Int16Array<ArrayBuffer>` | `__int16array.with` | 📋 Planned | - |
| `new Int16Array(elements: Iterable<number>): Int16Array<ArrayBuffer>` | `new (elements: Iterable<number>): Int16Array<ArrayBuffer>` | `__int16array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `int16array` are organized per API under `internal/compiler/testdata/corpus/int16array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/int16array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
