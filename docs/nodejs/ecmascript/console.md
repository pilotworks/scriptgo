# Console Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 Console Specification](https://tc39.es/ecma262/#sec-console-objects)  
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
| `Console.assert(condition?: boolean, ...data: any[]): void` | `assert(condition?: boolean, ...data: any[]): void` | `__console.assert` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.clear(): void` | `clear(): void` | `__console.clear` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.count(label?: string): void` | `count(label?: string): void` | `__console.count` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.countReset(label?: string): void` | `countReset(label?: string): void` | `__console.countReset` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.debug(...args: any[]): void` | `debug(...args: any[]): void` | `__console.debug` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.dir(item?: any, options?: any): void` | `dir(item?: any, options?: any): void` | `__console.dir` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.dirxml(...data: any[]): void` | `dirxml(...data: any[]): void` | `__console.dirxml` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.error(...args: any[]): void` | `error(...args: any[]): void` | `__console.error` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.group(...data: any[]): void` | `group(...data: any[]): void` | `__console.group` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.groupCollapsed(...data: any[]): void` | `groupCollapsed(...data: any[]): void` | `__console.groupCollapsed` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.groupEnd(): void` | `groupEnd(): void` | `__console.groupEnd` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.info(...args: any[]): void` | `info(...args: any[]): void` | `__console.info` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.log(...args: any[]): void` | `log(...args: any[]): void` | `__console.log` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.table(tabularData?: any, properties?: string[]): void` | `table(tabularData?: any, properties?: string[]): void` | `__console.table` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.time(label?: string): void` | `time(label?: string): void` | `__console.time` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.timeEnd(label?: string): void` | `timeEnd(label?: string): void` | `__console.timeEnd` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.timeLog(label?: string, ...data: any[]): void` | `timeLog(label?: string, ...data: any[]): void` | `__console.timeLog` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.trace(...data: any[]): void` | `trace(...data: any[]): void` | `__console.trace` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `Console.warn(...args: any[]): void` | `warn(...args: any[]): void` | `__console.warn` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |
| `new Console(): Console` | `new(): Console` | `__console.new` | ✅ Done | `internal/compiler/testdata/corpus/api/console.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `console` are organized per API under `internal/compiler/testdata/corpus/console/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/console/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
