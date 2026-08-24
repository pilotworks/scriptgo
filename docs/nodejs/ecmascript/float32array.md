# Float32Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Float32Array Specification](https://tc39.es/ecma262/#sec-float32array-objects)  
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
| `Float32Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__float32array.at` | 📋 Planned | - |
| `Float32Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__float32array.copyWithin` | 📋 Planned | - |
| `Float32Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__float32array.entries` | 📋 Planned | - |
| `Float32Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__float32array.every` | 📋 Planned | - |
| `Float32Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__float32array.fill` | 📋 Planned | - |
| `Float32Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Float32Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Float32Array<ArrayBuffer>` | `__float32array.filter` | 📋 Planned | - |
| `Float32Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__float32array.find` | 📋 Planned | - |
| `Float32Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__float32array.findIndex` | 📋 Planned | - |
| `Float32Array.findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__float32array.findLast` | 📋 Planned | - |
| `Float32Array.findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `__float32array.findLastIndex` | 📋 Planned | - |
| `Float32Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__float32array.forEach` | 📋 Planned | - |
| `Float32Array.from(elements: Iterable<number>): Float32Array<ArrayBuffer>` | `from(elements: Iterable<number>): Float32Array<ArrayBuffer>` | `__float32array.from` | 📋 Planned | - |
| `Float32Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__float32array.includes` | 📋 Planned | - |
| `Float32Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__float32array.indexOf` | 📋 Planned | - |
| `Float32Array.join(separator?: string): string` | `join(separator?: string): string` | `__float32array.join` | 📋 Planned | - |
| `Float32Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__float32array.keys` | 📋 Planned | - |
| `Float32Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__float32array.lastIndexOf` | 📋 Planned | - |
| `Float32Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Float32Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Float32Array<ArrayBuffer>` | `__float32array.map` | 📋 Planned | - |
| `Float32Array.of(...items: number[]): Float32Array<ArrayBuffer>` | `of(...items: number[]): Float32Array<ArrayBuffer>` | `__float32array.of` | 📋 Planned | - |
| `Float32Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__float32array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Float32Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__float32array.buffer` | 📋 Planned | - |
| `Float32Array.readonly byteLength: number` | `readonly byteLength: number` | `__float32array.byteLength` | 📋 Planned | - |
| `Float32Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__float32array.byteOffset` | 📋 Planned | - |
| `Float32Array.readonly length: number` | `readonly length: number` | `__float32array.length` | 📋 Planned | - |
| `Float32Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__float32array.reduce` | 📋 Planned | - |
| `Float32Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__float32array.reduceRight` | 📋 Planned | - |
| `Float32Array.reverse(): this` | `reverse(): this` | `__float32array.reverse` | 📋 Planned | - |
| `Float32Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__float32array.set` | 📋 Planned | - |
| `Float32Array.slice(start?: number, end?: number): Float32Array<ArrayBuffer>` | `slice(start?: number, end?: number): Float32Array<ArrayBuffer>` | `__float32array.slice` | 📋 Planned | - |
| `Float32Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__float32array.some` | 📋 Planned | - |
| `Float32Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__float32array.sort` | 📋 Planned | - |
| `Float32Array.subarray(begin?: number, end?: number): Float32Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Float32Array<TArrayBuffer>` | `__float32array.subarray` | 📋 Planned | - |
| `Float32Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__float32array.toLocaleString` | 📋 Planned | - |
| `Float32Array.toReversed(): Float32Array<ArrayBuffer>` | `toReversed(): Float32Array<ArrayBuffer>` | `__float32array.toReversed` | 📋 Planned | - |
| `Float32Array.toSorted(compareFn?: (a: number, b: number) => number): Float32Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Float32Array<ArrayBuffer>` | `__float32array.toSorted` | 📋 Planned | - |
| `Float32Array.toString(): string` | `toString(): string` | `__float32array.toString` | 📋 Planned | - |
| `Float32Array.valueOf(): this` | `valueOf(): this` | `__float32array.valueOf` | 📋 Planned | - |
| `Float32Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__float32array.values` | 📋 Planned | - |
| `Float32Array.with(index: number, value: number): Float32Array<ArrayBuffer>` | `with(index: number, value: number): Float32Array<ArrayBuffer>` | `__float32array.with` | 📋 Planned | - |
| `new Float32Array(elements: Iterable<number>): Float32Array<ArrayBuffer>` | `new (elements: Iterable<number>): Float32Array<ArrayBuffer>` | `__float32array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `float32array` are organized per API under `internal/compiler/testdata/corpus/float32array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/float32array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
