# BigInt64Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 BigInt64Array Specification](https://tc39.es/ecma262/#sec-bigint64array-objects)  
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
| `BigInt64Array.at(index: number): bigint \| undefined` | `at(index: number): bigint \| undefined` | `__bigint64array.at` | 📋 Planned | - |
| `BigInt64Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__bigint64array.copyWithin` | 📋 Planned | - |
| `BigInt64Array.entries(): ArrayIterator<[number, bigint]>` | `entries(): ArrayIterator<[number, bigint]>` | `__bigint64array.entries` | 📋 Planned | - |
| `BigInt64Array.every(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `every(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `__bigint64array.every` | 📋 Planned | - |
| `BigInt64Array.fill(value: bigint, start?: number, end?: number): this` | `fill(value: bigint, start?: number, end?: number): this` | `__bigint64array.fill` | 📋 Planned | - |
| `BigInt64Array.filter(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => any, thisArg?: any): BigInt64Array<ArrayBuffer>` | `filter(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => any, thisArg?: any): BigInt64Array<ArrayBuffer>` | `__bigint64array.filter` | 📋 Planned | - |
| `BigInt64Array.find(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): bigint \| undefined` | `find(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): bigint \| undefined` | `__bigint64array.find` | 📋 Planned | - |
| `BigInt64Array.findIndex(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): number` | `__bigint64array.findIndex` | 📋 Planned | - |
| `BigInt64Array.findLast<S extends bigint>( predicate: ( value: bigint, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends bigint>( predicate: ( value: bigint, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__bigint64array.findLast` | 📋 Planned | - |
| `BigInt64Array.findLastIndex( predicate: ( value: bigint, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: ( value: bigint, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `__bigint64array.findLastIndex` | 📋 Planned | - |
| `BigInt64Array.forEach(callbackfn: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => void, thisArg?: any): void` | `forEach(callbackfn: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => void, thisArg?: any): void` | `__bigint64array.forEach` | 📋 Planned | - |
| `BigInt64Array.from(arrayLike: ArrayLike<bigint>): BigInt64Array<ArrayBuffer>` | `from(arrayLike: ArrayLike<bigint>): BigInt64Array<ArrayBuffer>` | `__bigint64array.from` | 📋 Planned | - |
| `BigInt64Array.includes(searchElement: bigint, fromIndex?: number): boolean` | `includes(searchElement: bigint, fromIndex?: number): boolean` | `__bigint64array.includes` | 📋 Planned | - |
| `BigInt64Array.indexOf(searchElement: bigint, fromIndex?: number): number` | `indexOf(searchElement: bigint, fromIndex?: number): number` | `__bigint64array.indexOf` | 📋 Planned | - |
| `BigInt64Array.join(separator?: string): string` | `join(separator?: string): string` | `__bigint64array.join` | 📋 Planned | - |
| `BigInt64Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__bigint64array.keys` | 📋 Planned | - |
| `BigInt64Array.lastIndexOf(searchElement: bigint, fromIndex?: number): number` | `lastIndexOf(searchElement: bigint, fromIndex?: number): number` | `__bigint64array.lastIndexOf` | 📋 Planned | - |
| `BigInt64Array.map(callbackfn: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => bigint, thisArg?: any): BigInt64Array<ArrayBuffer>` | `map(callbackfn: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => bigint, thisArg?: any): BigInt64Array<ArrayBuffer>` | `__bigint64array.map` | 📋 Planned | - |
| `BigInt64Array.of(...items: bigint[]): BigInt64Array<ArrayBuffer>` | `of(...items: bigint[]): BigInt64Array<ArrayBuffer>` | `__bigint64array.of` | 📋 Planned | - |
| `BigInt64Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__bigint64array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `BigInt64Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__bigint64array.buffer` | 📋 Planned | - |
| `BigInt64Array.readonly byteLength: number` | `readonly byteLength: number` | `__bigint64array.byteLength` | 📋 Planned | - |
| `BigInt64Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__bigint64array.byteOffset` | 📋 Planned | - |
| `BigInt64Array.readonly length: number` | `readonly length: number` | `__bigint64array.length` | 📋 Planned | - |
| `BigInt64Array.reduce(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigInt64Array<TArrayBuffer>) => bigint): bigint` | `reduce(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigInt64Array<TArrayBuffer>) => bigint): bigint` | `__bigint64array.reduce` | 📋 Planned | - |
| `BigInt64Array.reduceRight(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigInt64Array<TArrayBuffer>) => bigint): bigint` | `reduceRight(callbackfn: (previousValue: bigint, currentValue: bigint, currentIndex: number, array: BigInt64Array<TArrayBuffer>) => bigint): bigint` | `__bigint64array.reduceRight` | 📋 Planned | - |
| `BigInt64Array.reverse(): this` | `reverse(): this` | `__bigint64array.reverse` | 📋 Planned | - |
| `BigInt64Array.set(array: ArrayLike<bigint>, offset?: number): void` | `set(array: ArrayLike<bigint>, offset?: number): void` | `__bigint64array.set` | 📋 Planned | - |
| `BigInt64Array.slice(start?: number, end?: number): BigInt64Array<ArrayBuffer>` | `slice(start?: number, end?: number): BigInt64Array<ArrayBuffer>` | `__bigint64array.slice` | 📋 Planned | - |
| `BigInt64Array.some(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `some(predicate: (value: bigint, index: number, array: BigInt64Array<TArrayBuffer>) => boolean, thisArg?: any): boolean` | `__bigint64array.some` | 📋 Planned | - |
| `BigInt64Array.sort(compareFn?: (a: bigint, b: bigint) => number \| bigint): this` | `sort(compareFn?: (a: bigint, b: bigint) => number \| bigint): this` | `__bigint64array.sort` | 📋 Planned | - |
| `BigInt64Array.subarray(begin?: number, end?: number): BigInt64Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): BigInt64Array<TArrayBuffer>` | `__bigint64array.subarray` | 📋 Planned | - |
| `BigInt64Array.toLocaleString(locales?: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales?: string \| string[], options?: Intl.NumberFormatOptions): string` | `__bigint64array.toLocaleString` | 📋 Planned | - |
| `BigInt64Array.toReversed(): BigInt64Array<ArrayBuffer>` | `toReversed(): BigInt64Array<ArrayBuffer>` | `__bigint64array.toReversed` | 📋 Planned | - |
| `BigInt64Array.toSorted(compareFn?: (a: bigint, b: bigint) => number): BigInt64Array<ArrayBuffer>` | `toSorted(compareFn?: (a: bigint, b: bigint) => number): BigInt64Array<ArrayBuffer>` | `__bigint64array.toSorted` | 📋 Planned | - |
| `BigInt64Array.toString(): string` | `toString(): string` | `__bigint64array.toString` | 📋 Planned | - |
| `BigInt64Array.valueOf(): BigInt64Array<TArrayBuffer>` | `valueOf(): BigInt64Array<TArrayBuffer>` | `__bigint64array.valueOf` | 📋 Planned | - |
| `BigInt64Array.values(): ArrayIterator<bigint>` | `values(): ArrayIterator<bigint>` | `__bigint64array.values` | 📋 Planned | - |
| `BigInt64Array.with(index: number, value: bigint): BigInt64Array<ArrayBuffer>` | `with(index: number, value: bigint): BigInt64Array<ArrayBuffer>` | `__bigint64array.with` | 📋 Planned | - |
| `new BigInt64Array(length?: number): BigInt64Array<ArrayBuffer>` | `new (length?: number): BigInt64Array<ArrayBuffer>` | `__bigint64array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `bigint64array` are organized per API under `internal/compiler/testdata/corpus/bigint64array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/bigint64array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
