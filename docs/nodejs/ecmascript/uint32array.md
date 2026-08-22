# Uint32Array Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Uint32Array Specification](https://tc39.es/ecma262/#sec-uint32array-objects)  
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
| `Uint32Array.at(index: number): number \| undefined` | `at(index: number): number \| undefined` | `__uint32array.at` | 📋 Planned | - |
| `Uint32Array.copyWithin(target: number, start: number, end?: number): this` | `copyWithin(target: number, start: number, end?: number): this` | `__uint32array.copyWithin` | 📋 Planned | - |
| `Uint32Array.entries(): ArrayIterator<[number, number]>` | `entries(): ArrayIterator<[number, number]>` | `__uint32array.entries` | 📋 Planned | - |
| `Uint32Array.every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `every(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint32array.every` | 📋 Planned | - |
| `Uint32Array.fill(value: number, start?: number, end?: number): this` | `fill(value: number, start?: number, end?: number): this` | `__uint32array.fill` | 📋 Planned | - |
| `Uint32Array.filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint32Array<ArrayBuffer>` | `filter(predicate: (value: number, index: number, array: this) => any, thisArg?: any): Uint32Array<ArrayBuffer>` | `__uint32array.filter` | 📋 Planned | - |
| `Uint32Array.find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `find(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number \| undefined` | `__uint32array.find` | 📋 Planned | - |
| `Uint32Array.findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `findIndex(predicate: (value: number, index: number, obj: this) => boolean, thisArg?: any): number` | `__uint32array.findIndex` | 📋 Planned | - |
| `Uint32Array.findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `findLast<S extends number>( predicate: ( value: number, index: number, array: this, ) => value is S, thisArg?: any, ): S \| undefined` | `__uint32array.findLast` | 📋 Planned | - |
| `Uint32Array.findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `findLastIndex( predicate: ( value: number, index: number, array: this, ) => unknown, thisArg?: any, ): number` | `__uint32array.findLastIndex` | 📋 Planned | - |
| `Uint32Array.forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `forEach(callbackfn: (value: number, index: number, array: this) => void, thisArg?: any): void` | `__uint32array.forEach` | 📋 Planned | - |
| `Uint32Array.from(elements: Iterable<number>): Uint32Array<ArrayBuffer>` | `from(elements: Iterable<number>): Uint32Array<ArrayBuffer>` | `__uint32array.from` | 📋 Planned | - |
| `Uint32Array.includes(searchElement: number, fromIndex?: number): boolean` | `includes(searchElement: number, fromIndex?: number): boolean` | `__uint32array.includes` | 📋 Planned | - |
| `Uint32Array.indexOf(searchElement: number, fromIndex?: number): number` | `indexOf(searchElement: number, fromIndex?: number): number` | `__uint32array.indexOf` | 📋 Planned | - |
| `Uint32Array.join(separator?: string): string` | `join(separator?: string): string` | `__uint32array.join` | 📋 Planned | - |
| `Uint32Array.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__uint32array.keys` | 📋 Planned | - |
| `Uint32Array.lastIndexOf(searchElement: number, fromIndex?: number): number` | `lastIndexOf(searchElement: number, fromIndex?: number): number` | `__uint32array.lastIndexOf` | 📋 Planned | - |
| `Uint32Array.map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint32Array<ArrayBuffer>` | `map(callbackfn: (value: number, index: number, array: this) => number, thisArg?: any): Uint32Array<ArrayBuffer>` | `__uint32array.map` | 📋 Planned | - |
| `Uint32Array.new <TArrayBuffer extends ArrayBufferLike = ArrayBuffer>(buffer: TArrayBuffer, byteOffset?: number, length?: number): Uint32Array<TArrayBuffer>` | `new <TArrayBuffer extends ArrayBufferLike = ArrayBuffer>(buffer: TArrayBuffer, byteOffset?: number, length?: number): Uint32Array<TArrayBuffer>` | `__uint32array.new` | 📋 Planned | - |
| `Uint32Array.of(...items: number[]): Uint32Array<ArrayBuffer>` | `of(...items: number[]): Uint32Array<ArrayBuffer>` | `__uint32array.of` | 📋 Planned | - |
| `Uint32Array.readonly BYTES_PER_ELEMENT: number` | `readonly BYTES_PER_ELEMENT: number` | `__uint32array.BYTES_PER_ELEMENT` | 📋 Planned | - |
| `Uint32Array.readonly buffer: TArrayBuffer` | `readonly buffer: TArrayBuffer` | `__uint32array.buffer` | 📋 Planned | - |
| `Uint32Array.readonly byteLength: number` | `readonly byteLength: number` | `__uint32array.byteLength` | 📋 Planned | - |
| `Uint32Array.readonly byteOffset: number` | `readonly byteOffset: number` | `__uint32array.byteOffset` | 📋 Planned | - |
| `Uint32Array.readonly length: number` | `readonly length: number` | `__uint32array.length` | 📋 Planned | - |
| `Uint32Array.reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduce(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint32array.reduce` | 📋 Planned | - |
| `Uint32Array.reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `reduceRight(callbackfn: (previousValue: number, currentValue: number, currentIndex: number, array: this) => number): number` | `__uint32array.reduceRight` | 📋 Planned | - |
| `Uint32Array.reverse(): this` | `reverse(): this` | `__uint32array.reverse` | 📋 Planned | - |
| `Uint32Array.set(array: ArrayLike<number>, offset?: number): void` | `set(array: ArrayLike<number>, offset?: number): void` | `__uint32array.set` | 📋 Planned | - |
| `Uint32Array.slice(start?: number, end?: number): Uint32Array<ArrayBuffer>` | `slice(start?: number, end?: number): Uint32Array<ArrayBuffer>` | `__uint32array.slice` | 📋 Planned | - |
| `Uint32Array.some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `some(predicate: (value: number, index: number, array: this) => unknown, thisArg?: any): boolean` | `__uint32array.some` | 📋 Planned | - |
| `Uint32Array.sort(compareFn?: (a: number, b: number) => number): this` | `sort(compareFn?: (a: number, b: number) => number): this` | `__uint32array.sort` | 📋 Planned | - |
| `Uint32Array.subarray(begin?: number, end?: number): Uint32Array<TArrayBuffer>` | `subarray(begin?: number, end?: number): Uint32Array<TArrayBuffer>` | `__uint32array.subarray` | 📋 Planned | - |
| `Uint32Array.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions): string` | `__uint32array.toLocaleString` | 📋 Planned | - |
| `Uint32Array.toReversed(): Uint32Array<ArrayBuffer>` | `toReversed(): Uint32Array<ArrayBuffer>` | `__uint32array.toReversed` | 📋 Planned | - |
| `Uint32Array.toSorted(compareFn?: (a: number, b: number) => number): Uint32Array<ArrayBuffer>` | `toSorted(compareFn?: (a: number, b: number) => number): Uint32Array<ArrayBuffer>` | `__uint32array.toSorted` | 📋 Planned | - |
| `Uint32Array.toString(): string` | `toString(): string` | `__uint32array.toString` | 📋 Planned | - |
| `Uint32Array.valueOf(): this` | `valueOf(): this` | `__uint32array.valueOf` | 📋 Planned | - |
| `Uint32Array.values(): ArrayIterator<number>` | `values(): ArrayIterator<number>` | `__uint32array.values` | 📋 Planned | - |
| `Uint32Array.with(index: number, value: number): Uint32Array<ArrayBuffer>` | `with(index: number, value: number): Uint32Array<ArrayBuffer>` | `__uint32array.with` | 📋 Planned | - |
| `new Uint32Array(elements: Iterable<number>): Uint32Array<ArrayBuffer>` | `new (elements: Iterable<number>): Uint32Array<ArrayBuffer>` | `__uint32array.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `uint32array` are organized per API under `internal/compiler/testdata/corpus/uint32array/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/uint32array/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
