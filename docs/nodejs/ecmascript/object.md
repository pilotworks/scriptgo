# Object Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Object Specification](https://tc39.es/ecma262/#sec-object-objects)  
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
| `Object.assign<T extends {}, U>(target: T, source: U): T & U` | `assign<T extends {}, U>(target: T, source: U): T & U` | `__object.assign` | ✅ Done | `internal/compiler/testdata/corpus/api/object.ts` |
| `Object.hasOwn(o: object, v: PropertyKey): boolean` | `hasOwn(o: object, v: PropertyKey): boolean` | `__object.hasOwn` | ✅ Done | `internal/compiler/testdata/corpus/api/object.ts` |
| `Object.is(value1: any, value2: any): boolean` | `is(value1: any, value2: any): boolean` | `__object.is` | ✅ Done | `internal/compiler/testdata/corpus/api/object.ts` |
| `Object.keys(o: {}): string[]` | `keys(o: {}): string[]` | `__object.keys` | ✅ Done | `internal/compiler/testdata/corpus/api/object.ts` |
| `Object.values<T>(o: { [s: string]: T; } \| ArrayLike<T>): T[]` | `values<T>(o: { [s: string]: T; } \| ArrayLike<T>): T[]` | `__object.values` | ✅ Done | `internal/compiler/testdata/corpus/api/object.ts` |
| `Object.create(o: object \| null): any` | `create(o: object \| null): any` | `__object.create` | 📋 Planned | - |
| `Object.defineProperties<T>(o: T, properties: PropertyDescriptorMap & ThisType<any>): T` | `defineProperties<T>(o: T, properties: PropertyDescriptorMap & ThisType<any>): T` | `__object.defineProperties` | 📋 Planned | - |
| `Object.defineProperty<T>(o: T, p: PropertyKey, attributes: PropertyDescriptor & ThisType<any>): T` | `defineProperty<T>(o: T, p: PropertyKey, attributes: PropertyDescriptor & ThisType<any>): T` | `__object.defineProperty` | 📋 Planned | - |
| `Object.entries<T>(o: { [s: string]: T; } \| ArrayLike<T>): [string, T][]` | `entries<T>(o: { [s: string]: T; } \| ArrayLike<T>): [string, T][]` | `__object.entries` | 📋 Planned | - |
| `Object.freeze<T extends Function>(f: T): T` | `freeze<T extends Function>(f: T): T` | `__object.freeze` | 📋 Planned | - |
| `Object.fromEntries<T = any>(entries: Iterable<readonly [PropertyKey, T]>): { [k: string]: T; }` | `fromEntries<T = any>(entries: Iterable<readonly [PropertyKey, T]>): { [k: string]: T; }` | `__object.fromEntries` | 📋 Planned | - |
| `Object.getOwnPropertyDescriptor(o: any, p: PropertyKey): PropertyDescriptor \| undefined` | `getOwnPropertyDescriptor(o: any, p: PropertyKey): PropertyDescriptor \| undefined` | `__object.getOwnPropertyDescriptor` | 📋 Planned | - |
| `Object.getOwnPropertyDescriptors<T>(o: T): { [P in keyof T]: TypedPropertyDescriptor<T[P]>; } & { [x: string]: PropertyDescriptor; }` | `getOwnPropertyDescriptors<T>(o: T): { [P in keyof T]: TypedPropertyDescriptor<T[P]>; } & { [x: string]: PropertyDescriptor; }` | `__object.getOwnPropertyDescriptors` | 📋 Planned | - |
| `Object.getOwnPropertyNames(o: any): string[]` | `getOwnPropertyNames(o: any): string[]` | `__object.getOwnPropertyNames` | 📋 Planned | - |
| `Object.getOwnPropertySymbols(o: any): symbol[]` | `getOwnPropertySymbols(o: any): symbol[]` | `__object.getOwnPropertySymbols` | 📋 Planned | - |
| `Object.getPrototypeOf(o: any): any` | `getPrototypeOf(o: any): any` | `__object.getPrototypeOf` | 📋 Planned | - |
| `Object.groupBy<K extends PropertyKey, T>( items: Iterable<T>, keySelector: (item: T, index: number) => K, ): Partial<Record<K, T[]>>` | `groupBy<K extends PropertyKey, T>( items: Iterable<T>, keySelector: (item: T, index: number) => K, ): Partial<Record<K, T[]>>` | `__object.groupBy` | 📋 Planned | - |
| `Object.hasOwnProperty(v: PropertyKey): boolean` | `hasOwnProperty(v: PropertyKey): boolean` | `__object.hasOwnProperty` | 📋 Planned | - |
| `Object.isExtensible(o: any): boolean` | `isExtensible(o: any): boolean` | `__object.isExtensible` | 📋 Planned | - |
| `Object.isFrozen(o: any): boolean` | `isFrozen(o: any): boolean` | `__object.isFrozen` | 📋 Planned | - |
| `Object.isPrototypeOf(v: Object): boolean` | `isPrototypeOf(v: Object): boolean` | `__object.isPrototypeOf` | 📋 Planned | - |
| `Object.isSealed(o: any): boolean` | `isSealed(o: any): boolean` | `__object.isSealed` | 📋 Planned | - |
| `Object.preventExtensions<T>(o: T): T` | `preventExtensions<T>(o: T): T` | `__object.preventExtensions` | 📋 Planned | - |
| `Object.propertyIsEnumerable(v: PropertyKey): boolean` | `propertyIsEnumerable(v: PropertyKey): boolean` | `__object.propertyIsEnumerable` | 📋 Planned | - |
| `Object.seal<T>(o: T): T` | `seal<T>(o: T): T` | `__object.seal` | 📋 Planned | - |
| `Object.setPrototypeOf(o: any, proto: object \| null): any` | `setPrototypeOf(o: any, proto: object \| null): any` | `__object.setPrototypeOf` | 📋 Planned | - |
| `Object.toLocaleString(): string` | `toLocaleString(): string` | `__object.toLocaleString` | 📋 Planned | - |
| `Object.toString(): string` | `toString(): string` | `__object.toString` | 📋 Planned | - |
| `Object.valueOf(): Object` | `valueOf(): Object` | `__object.valueOf` | 📋 Planned | - |
| `new Objectconstructor: Function` | `constructor: Function` | `__object.new` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `object` are organized per API under `internal/compiler/testdata/corpus/object/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/object/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
