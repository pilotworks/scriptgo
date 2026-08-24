# BigUint64Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 BigUint64Array Specification](https://tc39.es/ecma262/#sec-biguint64array-objects)  
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
| `BigUint64Array.at(index: number): bigint \| undefined` | `at(index: number): bigint \| undefined` | `__biguint64array.at` | 📋 Planned | - |
| `BigUint64Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__biguint64array.copyWithin` | 📋 Planned | - |
| `BigUint64Array.entries(): ArrayIterator<[number, bigint]>` | `entries(): ArrayIterator<[number, bigint]>` | `__biguint64array.entries` | 📋 Planned | - |
| `BigUint64Array.every(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `every(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `__biguint64array.every` | 📋 Planned | - |
| `BigUint64Array.fill(value: bigint, start?: number, end?: number): this` | `fill(value: bigint, start?: number, end?: number): this` | `__biguint64array.fill` | 📋 Planned | - |
| `BigUint64Array.filter(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => any, thisArg?: any): BigUint64Array<ArrayBuffer>` | `filter(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => any, thisArg?: any): BigUint64Array<ArrayBuffer>` | `__biguint64array.filter` | 📋 Planned | - |
| `BigUint64Array.find(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): bigint \| undefined` | `find(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): bigint \| undefined` | `__biguint64array.find` | 📋 Planned | - |
| `BigUint64Array.findIndex(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): number` | `__biguint64array.findIndex` | 📋 Planned | - |
| `BigUint64Array.findLast<S extends bigint>( predicate: ( value: bigint, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends bigint>( predicate: ( value: bigint, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__biguint64array.findLast` | 📋 Planned | - |
| `BigUint64Array.findLastIndex( predicate: ( value: bigint, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: ( value: bigint, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `__biguint64array.findLastIndex` | 📋 Planned | - |
| `BigUint64Array.forEach(callbackfn: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => void, thisArg?: any): void` | `forEach(callbackfn: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => void, thisArg?: any): void` | `__biguint64array.forEach` | 📋 Planned | - |
| `BigUint64Array.from(arrayLike: ArrayLike<bigint>): BigUint64Array<ArrayBuffer>` | `from(arrayLike: ArrayLike<bigint>): BigUint64Array<ArrayBuffer>` | `__biguint64array.from` | 📋 Planned | - |
| `BigUint64Array.includes(searchElement: bigint, fromIndex?: number): boolean` | `includes(searchElement: bigint, fromIndex?: number): boolean` | `__biguint64array.includes` | 📋 Planned | - |
| `BigUint64Array.indexOf(searchElement: bigint, fromIndex?: number): number` | `indexOf(searchElement: bigint, fromIndex?: number): number` | `__biguint64array.indexOf` | 📋 Planned | - |
| `BigUint64Array.join(separator?: string): string` | `join(separator?: string): string` | `__biguint64array.join` | 📋 Planned | - |
| `BigUint64Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__biguint64array.keys` | 📋 Planned | - |
| `BigUint64Array.lastIndexOf(searchElement: bigint, fromIndex?: number): number` | `lastIndexOf(searchElement: bigint, fromIndex?: number): number` | `__biguint64array.lastIndexOf` | 📋 Planned | - |
| `BigUint64Array.map(callbackfn: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => bigint, thisArg?: any): BigUint64Array<ArrayBuffer>` | `map(callbackfn: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => bigint, thisArg?: any): BigUint64Array<ArrayBuffer>` | `__biguint64array.map` | 📋 Planned | - |
| `BigUint64Array.of(...items: bigint[]): BigUint64Array<ArrayBuffer>` | `of(...items: bigint[]): BigUint64Array<ArrayBuffer>` | `__biguint64array.of` | 📋 Planned | - |
| `BigUint64Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__biguint64array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `BigUint64Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__biguint64array.buffer` | 📋 Planned | - |
| `BigUint64Array.readonly byteLength: number` | `readonly byteLength: number` | `__biguint64array.byteLength` | 📋 Planned | - |
| `BigUint64Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__biguint64array.byteOffset` | 📋 Planned | - |
| `BigUint64Array.readonly length: number` | `readonly length: number` | `__biguint64array.length` | 📋 Planned | - |
| `BigUint64Array.reduce(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigUint64Array<TArrayBuffer>) => bigint): bigint` | `reduce(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigUint64Array<TArrayBuffer>) => bigint): bigint` | `__biguint64array.reduce` | 📋 Planned | - |
| `BigUint64Array.reduceRight(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigUint64Array<TArrayBuffer>) => bigint): bigint` | `reduceRight(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigUint64Array<TArrayBuffer>) => bigint): bigint` | `__biguint64array.reduceRight` | 📋 Planned | - |
| `BigUint64Array.reverse(): this` | `reverse(): this` | `__biguint64array.reverse` | 📋 Planned | - |
| `BigUint64Array.set(array: ArrayLike<bigint>, offset?: number): void` | `set(array: ArrayLike<bigint>, offset?: number): void` | `__biguint64array.set` | 📋 Planned | - |
| `BigUint64Array.slice(start?: number, end?: number): BigUint64Array<ArrayBuffer>` | `slice(start?: number, end?: number): BigUint64Array<ArrayBuffer>` | `__biguint64array.slice` | 📋 Planned | - |
| `BigUint64Array.some(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `some(predicate: (value: bigint, index: number, array: BigUint64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `__biguint64array.some` | 📋 Planned | - |
| `BigUint64Array.sort(compareFn?: (a: bigint, b: bigint) => number \| bigint): this` | `sort(compareFn?: (a: bigint, b: bigint) => number \| bigint): this` | `__biguint64array.sort` | 📋 Planned | - |
| `BigUint64Array.subarray(begin?: number, end?: number): BigUint64Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): BigUint64Array<TArrayBuffer>` | `__biguint64array.subarray` | 📋 Planned | - |
| `BigUint64Array.toLocaleString(locales?: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales?: string \| string[], options?: Intl.NumberFormatOptions): string` | `__biguint64array.toLocaleString` | 📋 Planned | - |
| `BigUint64Array.toReversed(): BigUint64Array<ArrayBuffer>` | `toReversed(): BigUint64Array<ArrayBuffer>` | `__biguint64array.toReversed` | 📋 Planned | - |
| `BigUint64Array.toSorted(compareFn?: (a: bigint, b: bigint) => number): BigUint64Array<ArrayBuffer>` | `toSorted(compareFn?: (a: bigint, b: bigint) => number): BigUint64Array<ArrayBuffer>` | `__biguint64array.toSorted` | 📋 Planned | - |
| `BigUint64Array.toString(): string` | `toString(): string` | `__biguint64array.toString` | 📋 Planned | - |
| `BigUint64Array.valueOf(): BigUint64Array<TArrayBuffer>` | `valueOf(): BigUint64Array<TArrayBuffer>` | `__biguint64array.valueOf` | 📋 Planned | - |
| `BigUint64Array.values(): ArrayIterator<bigint>` | `values(): ArrayIterator<bigint>` | `__biguint64array.values` | 📋 Planned | - |
| `BigUint64Array.with(index: number, value: bigint): BigUint64Array<ArrayBuffer>` | `with(index: number, value: bigint): BigUint64Array<ArrayBuffer>` | `__biguint64array.with` | 📋 Planned | - |
| `new BigUint64Array(length?: number): BigUint64Array<ArrayBuffer>` | `new (length?: number): BigUint64Array<ArrayBuffer>` | `__biguint64array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `biguint64array` are organized per API under `internal/compiler/testdata/corpus/biguint64array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/biguint64array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
