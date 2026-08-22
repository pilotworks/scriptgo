# Float64Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Float64Array Specification](https://tc39.es/ecma262/#sec-float64array-objects)  
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
| `Float64Array.from(elements: Iterable<number>): Float64Array<ArrayBuffer>` | `from(elements: Iterable<number>): Float64Array<ArrayBuffer>` | `__float64array.from` | ✅ Done | `internal/compiler/testdata/corpus/api/float64array.ts` |
| `Float64Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__float64array.at` | 📋 Planned | - |
| `Float64Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__float64array.copyWithin` | 📋 Planned | - |
| `Float64Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__float64array.entries` | 📋 Planned | - |
| `Float64Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__float64array.every` | 📋 Planned | - |
| `Float64Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__float64array.fill` | 📋 Planned | - |
| `Float64Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Float64Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Float64Array<ArrayBuffer>` | `__float64array.filter` | 📋 Planned | - |
| `Float64Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__float64array.find` | 📋 Planned | - |
| `Float64Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__float64array.findIndex` | 📋 Planned | - |
| `Float64Array.findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__float64array.findLast` | 📋 Planned | - |
| `Float64Array.findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `__float64array.findLastIndex` | 📋 Planned | - |
| `Float64Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__float64array.forEach` | 📋 Planned | - |
| `Float64Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__float64array.includes` | 📋 Planned | - |
| `Float64Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__float64array.indexOf` | 📋 Planned | - |
| `Float64Array.join(separator?: string): string` | `join(separator?: string): string` | `__float64array.join` | 📋 Planned | - |
| `Float64Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__float64array.keys` | 📋 Planned | - |
| `Float64Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__float64array.lastIndexOf` | 📋 Planned | - |
| `Float64Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Float64Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Float64Array<ArrayBuffer>` | `__float64array.map` | 📋 Planned | - |
| `Float64Array.of(...items: number[]): Float64Array<ArrayBuffer>` | `of(...items: number[]): Float64Array<ArrayBuffer>` | `__float64array.of` | 📋 Planned | - |
| `Float64Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__float64array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Float64Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__float64array.buffer` | 📋 Planned | - |
| `Float64Array.readonly byteLength: number` | `readonly byteLength: number` | `__float64array.byteLength` | 📋 Planned | - |
| `Float64Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__float64array.byteOffset` | 📋 Planned | - |
| `Float64Array.readonly length: number` | `readonly length: number` | `__float64array.length` | 📋 Planned | - |
| `Float64Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__float64array.reduce` | 📋 Planned | - |
| `Float64Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__float64array.reduceRight` | 📋 Planned | - |
| `Float64Array.reverse(): this` | `reverse(): this` | `__float64array.reverse` | 📋 Planned | - |
| `Float64Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__float64array.set` | 📋 Planned | - |
| `Float64Array.slice(start?: number, end?: number): Float64Array<ArrayBuffer>` | `slice(start?: number, end?: number): Float64Array<ArrayBuffer>` | `__float64array.slice` | 📋 Planned | - |
| `Float64Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__float64array.some` | 📋 Planned | - |
| `Float64Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__float64array.sort` | 📋 Planned | - |
| `Float64Array.subarray(begin?: number, end?: number): Float64Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Float64Array<TArrayBuffer>` | `__float64array.subarray` | 📋 Planned | - |
| `Float64Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__float64array.toLocaleString` | 📋 Planned | - |
| `Float64Array.toReversed(): Float64Array<ArrayBuffer>` | `toReversed(): Float64Array<ArrayBuffer>` | `__float64array.toReversed` | 📋 Planned | - |
| `Float64Array.toSorted(compareFn?: (a: number, b: number) => number): Float64Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Float64Array<ArrayBuffer>` | `__float64array.toSorted` | 📋 Planned | - |
| `Float64Array.toString(): string` | `toString(): string` | `__float64array.toString` | 📋 Planned | - |
| `Float64Array.valueOf(): this` | `valueOf(): this` | `__float64array.valueOf` | 📋 Planned | - |
| `Float64Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__float64array.values` | 📋 Planned | - |
| `Float64Array.with(index: number, value: number): Float64Array<ArrayBuffer>` | `with(index: number, value: number): Float64Array<ArrayBuffer>` | `__float64array.with` | 📋 Planned | - |
| `new Float64Array(elements: Iterable<number>): Float64Array<ArrayBuffer>` | `new (elements: Iterable<number>): Float64Array<ArrayBuffer>` | `__float64array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `float64array` are organized per API under `internal/compiler/testdata/corpus/float64array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/float64array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
