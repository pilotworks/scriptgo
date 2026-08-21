# ReadonlyArray Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 ReadonlyArray Specification](https://tc39.es/ecma262/#sec-readonlyarray-objects)  
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
| `ReadonlyArray.): U[]` | `): U[]` | `__readonlyarray.)` | 📋 Planned | - |
| `ReadonlyArray.at(index: number): T \| undefined` | `at(index: number): T \| undefined` | `__readonlyarray.at` | 📋 Planned | - |
| `ReadonlyArray.callback: (this: This, value: T, index: number, array: T[]) => U \| ReadonlyArray<U>,` | `callback: (this: This, value: T, index: number, array: T[]) => U \| ReadonlyArray<U>,` | `__readonlyarray.callback` | 📋 Planned | - |
| `ReadonlyArray.concat(...items: ConcatArray<T>[]): T[]` | `concat(...items: ConcatArray<T>[]): T[]` | `__readonlyarray.concat` | 📋 Planned | - |
| `ReadonlyArray.depth?: D,` | `depth?: D,` | `__readonlyarray.depth?` | 📋 Planned | - |
| `ReadonlyArray.entries(): ArrayIterator<[number, T]>` | `entries(): ArrayIterator<[number, T]>` | `__readonlyarray.entries` | 📋 Planned | - |
| `ReadonlyArray.every(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): boolean` | `every(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): boolean` | `__readonlyarray.every` | 📋 Planned | - |
| `ReadonlyArray.filter(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): T[]` | `filter(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): T[]` | `__readonlyarray.filter` | 📋 Planned | - |
| `ReadonlyArray.find(predicate: (value: T, index: number, obj: readonly T[]) => unknown, thisArg?: any): T \| undefined` | `find(predicate: (value: T, index: number, obj: readonly T[]) => unknown, thisArg?: any): T \| undefined` | `__readonlyarray.find` | 📋 Planned | - |
| `ReadonlyArray.findIndex(predicate: (value: T, index: number, obj: readonly T[]) => unknown, thisArg?: any): number` | `findIndex(predicate: (value: T, index: number, obj: readonly T[]) => unknown, thisArg?: any): number` | `__readonlyarray.findIndex` | 📋 Planned | - |
| `ReadonlyArray.findLast(` | `findLast(` | `__readonlyarray.findLast` | 📋 Planned | - |
| `ReadonlyArray.findLastIndex(` | `findLastIndex(` | `__readonlyarray.findLastIndex` | 📋 Planned | - |
| `ReadonlyArray.forEach(callbackfn: (value: T, index: number, array: readonly T[]) => void, thisArg?: any): void` | `forEach(callbackfn: (value: T, index: number, array: readonly T[]) => void, thisArg?: any): void` | `__readonlyarray.forEach` | 📋 Planned | - |
| `ReadonlyArray.includes(searchElement: T, fromIndex?: number): boolean` | `includes(searchElement: T, fromIndex?: number): boolean` | `__readonlyarray.includes` | 📋 Planned | - |
| `ReadonlyArray.indexOf(searchElement: T, fromIndex?: number): number` | `indexOf(searchElement: T, fromIndex?: number): number` | `__readonlyarray.indexOf` | 📋 Planned | - |
| `ReadonlyArray.join(separator?: string): string` | `join(separator?: string): string` | `__readonlyarray.join` | 📋 Planned | - |
| `ReadonlyArray.keys(): ArrayIterator<number>` | `keys(): ArrayIterator<number>` | `__readonlyarray.keys` | 📋 Planned | - |
| `ReadonlyArray.lastIndexOf(searchElement: T, fromIndex?: number): number` | `lastIndexOf(searchElement: T, fromIndex?: number): number` | `__readonlyarray.lastIndexOf` | 📋 Planned | - |
| `ReadonlyArray.map<U>(callbackfn: (value: T, index: number, array: readonly T[]) => U, thisArg?: any): U[]` | `map<U>(callbackfn: (value: T, index: number, array: readonly T[]) => U, thisArg?: any): U[]` | `__readonlyarray.map<U>` | 📋 Planned | - |
| `ReadonlyArray.predicate: (value: T, index: number, array: readonly T[]) => value is S,` | `predicate: (value: T, index: number, array: readonly T[]) => value is S,` | `__readonlyarray.predicate` | 📋 Planned | - |
| `ReadonlyArray.readonly length: number` | `readonly length: number` | `__readonlyarray.length` | 📋 Planned | - |
| `ReadonlyArray.reduce(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: readonly T[]) => T): T` | `reduce(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: readonly T[]) => T): T` | `__readonlyarray.reduce` | 📋 Planned | - |
| `ReadonlyArray.reduce<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: readonly T[]) => U, initialValue: U): U` | `reduce<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: readonly T[]) => U, initialValue: U): U` | `__readonlyarray.reduce<U>` | 📋 Planned | - |
| `ReadonlyArray.reduceRight(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: readonly T[]) => T): T` | `reduceRight(callbackfn: (previousValue: T, currentValue: T, currentIndex: number, array: readonly T[]) => T): T` | `__readonlyarray.reduceRight` | 📋 Planned | - |
| `ReadonlyArray.reduceRight<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: readonly T[]) => U, initialValue: U): U` | `reduceRight<U>(callbackfn: (previousValue: U, currentValue: T, currentIndex: number, array: readonly T[]) => U, initialValue: U): U` | `__readonlyarray.reduceRight<U>` | 📋 Planned | - |
| `ReadonlyArray.slice(start?: number, end?: number): T[]` | `slice(start?: number, end?: number): T[]` | `__readonlyarray.slice` | 📋 Planned | - |
| `ReadonlyArray.some(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): boolean` | `some(predicate: (value: T, index: number, array: readonly T[]) => unknown, thisArg?: any): boolean` | `__readonlyarray.some` | 📋 Planned | - |
| `ReadonlyArray.this: A,` | `this: A,` | `__readonlyarray.this` | 📋 Planned | - |
| `ReadonlyArray.thisArg?: This,` | `thisArg?: This,` | `__readonlyarray.thisArg?` | 📋 Planned | - |
| `ReadonlyArray.toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions & Intl.DateTimeFormatOptions): string` | `toLocaleString(locales: string \| string[], options?: Intl.NumberFormatOptions & Intl.DateTimeFormatOptions): string` | `__readonlyarray.toLocaleString` | 📋 Planned | - |
| `ReadonlyArray.toReversed(): T[]` | `toReversed(): T[]` | `__readonlyarray.toReversed` | 📋 Planned | - |
| `ReadonlyArray.toSorted(compareFn?: (a: T, b: T) => number): T[]` | `toSorted(compareFn?: (a: T, b: T) => number): T[]` | `__readonlyarray.toSorted` | 📋 Planned | - |
| `ReadonlyArray.toSpliced(start: number, deleteCount: number, ...items: T[]): T[]` | `toSpliced(start: number, deleteCount: number, ...items: T[]): T[]` | `__readonlyarray.toSpliced` | 📋 Planned | - |
| `ReadonlyArray.toString(): string` | `toString(): string` | `__readonlyarray.toString` | 📋 Planned | - |
| `ReadonlyArray.values(): ArrayIterator<T>` | `values(): ArrayIterator<T>` | `__readonlyarray.values` | 📋 Planned | - |
| `ReadonlyArray.with(index: number, value: T): T[]` | `with(index: number, value: T): T[]` | `__readonlyarray.with` | 📋 Planned | - |
| `ReadonlyArray.}` | `}` | `__readonlyarray.}` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `readonlyarray` are organized per API under `internal/compiler/testdata/corpus/readonlyarray/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/readonlyarray/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
