# Reflect Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Reflect Specification](https://tc39.es/ecma262/#sec-reflect-objects)  
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
| `Reflect.apply<T, A extends readonly any[], R>( target: (this: T, ...args: A) => R, thisArgument: T, argumentsList: Readonly<A>, ): R` | `apply<T, A extends readonly any[], R>( target: (this: T, ...args: A) => R, thisArgument: T, argumentsList: Readonly<A>, ): R` | `__reflect.apply` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.construct<A extends readonly any[], R>( target: new (...args: A) => R, argumentsList: Readonly<A>, newTarget?: new (...args: any) => any, ): R` | `construct<A extends readonly any[], R>( target: new (...args: A) => R, argumentsList: Readonly<A>, newTarget?: new (...args: any) => any, ): R` | `__reflect.construct` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.defineMetadata(metadataKey: any, metadataValue: any, target: any, propertyKey?: string \| symbol): void` | `defineMetadata(metadataKey: any, metadataValue: any, target: any, propertyKey?: string \| symbol): void` | `__reflect.defineMetadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.defineProperty(target: object, propertyKey: PropertyKey, attributes: PropertyDescriptor & ThisType<any>): boolean` | `defineProperty(target: object, propertyKey: PropertyKey, attributes: PropertyDescriptor & ThisType<any>): boolean` | `__reflect.defineProperty` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.deleteProperty(target: object, propertyKey: PropertyKey): boolean` | `deleteProperty(target: object, propertyKey: PropertyKey): boolean` | `__reflect.deleteProperty` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.get<T extends object, P extends PropertyKey>( target: T, propertyKey: P, receiver?: unknown, ): P extends keyof T ? T[P] : any` | `get<T extends object, P extends PropertyKey>( target: T, propertyKey: P, receiver?: unknown, ): P extends keyof T ? T[P] : any` | `__reflect.get` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.getMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): any` | `getMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): any` | `__reflect.getMetadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.getOwnMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): any` | `getOwnMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): any` | `__reflect.getOwnMetadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.getOwnPropertyDescriptor<T extends object, P extends PropertyKey>( target: T, propertyKey: P, ): TypedPropertyDescriptor<P extends keyof T ? T[P] : any> \| undefined` | `getOwnPropertyDescriptor<T extends object, P extends PropertyKey>( target: T, propertyKey: P, ): TypedPropertyDescriptor<P extends keyof T ? T[P] : any> \| undefined` | `__reflect.getOwnPropertyDescriptor` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.getPrototypeOf(target: object): object \| null` | `getPrototypeOf(target: object): object \| null` | `__reflect.getPrototypeOf` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.has(target: object, propertyKey: PropertyKey): boolean` | `has(target: object, propertyKey: PropertyKey): boolean` | `__reflect.has` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.hasMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): boolean` | `hasMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): boolean` | `__reflect.hasMetadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.hasOwnMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): boolean` | `hasOwnMetadata(metadataKey: any, target: any, propertyKey?: string \| symbol): boolean` | `__reflect.hasOwnMetadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.isExtensible(target: object): boolean` | `isExtensible(target: object): boolean` | `__reflect.isExtensible` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.metadata(metadataKey: any, metadataValue: any): (target: any, propertyKey?: any) => void` | `metadata(metadataKey: any, metadataValue: any): (target: any, propertyKey?: any) => void` | `__reflect.metadata` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.ownKeys(target: object): (string \| symbol)[]` | `ownKeys(target: object): (string \| symbol)[]` | `__reflect.ownKeys` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.preventExtensions(target: object): boolean` | `preventExtensions(target: object): boolean` | `__reflect.preventExtensions` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.set<T extends object, P extends PropertyKey>( target: T, propertyKey: P, value: P extends keyof T ? T[P] : any, receiver?: any, ): boolean` | `set<T extends object, P extends PropertyKey>( target: T, propertyKey: P, value: P extends keyof T ? T[P] : any, receiver?: any, ): boolean` | `__reflect.set` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |
| `Reflect.setPrototypeOf(target: object, proto: object \| null): boolean` | `setPrototypeOf(target: object, proto: object \| null): boolean` | `__reflect.setPrototypeOf` | ✅ Done | `internal/compiler/testdata/corpus/api/reflect.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `reflect` are organized per API under `internal/compiler/testdata/corpus/reflect/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/reflect/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
