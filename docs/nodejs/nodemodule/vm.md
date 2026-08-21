# vm Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:vm`  
> **Specification Reference**: [Node.js 22 LTS vm Documentation](https://nodejs.org/docs/latest-v22.x/api/vm.html)  
> **Type Definition Source**: [@types/node/vm.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-vm-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:vm`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `cachedDataRejected` | `any` | `__vm.cachedDataRejected` | 📋 Planned | - |
| `constants` | `any` | `__vm.constants` | 📋 Planned | - |
| `dependencySpecifiers` {string\[]}` | `any` | `__vm.dependencySpecifiers` {string\[]}` | 📋 Planned | - |
| `error` | `any` | `__vm.error` | 📋 Planned | - |
| `identifier` | `any` | `__vm.identifier` | 📋 Planned | - |
| `module.evaluate([options])` | `(...) => any` | `__vm.module.evaluate` | 📋 Planned | - |
| `module.link(linker)` | `(...) => any` | `__vm.module.link` | 📋 Planned | - |
| `moduleRequests` {ModuleRequest\[]} Dependencies of this module.` | `any` | `__vm.moduleRequests` {ModuleRequest\[]} Dependencies of this module.` | 📋 Planned | - |
| `namespace` | `any` | `__vm.namespace` | 📋 Planned | - |
| `script.createCachedData()` | `(...) => any` | `__vm.script.createCachedData` | 📋 Planned | - |
| `script.runInContext(contextifiedObject[, options])` | `(...) => any` | `__vm.script.runInContext` | 📋 Planned | - |
| `script.runInNewContext([contextObject[, options]])` | `(...) => any` | `__vm.script.runInNewContext` | 📋 Planned | - |
| `script.runInThisContext([options])` | `(...) => any` | `__vm.script.runInThisContext` | 📋 Planned | - |
| `sourceMapURL` | `any` | `__vm.sourceMapURL` | 📋 Planned | - |
| `sourceTextModule.createCachedData()` | `(...) => any` | `__vm.sourceTextModule.createCachedData` | 📋 Planned | - |
| `sourceTextModule.instantiate()` | `(...) => any` | `__vm.sourceTextModule.instantiate` | 📋 Planned | - |
| `sourceTextModule.linkRequests(modules)` | `(...) => any` | `__vm.sourceTextModule.linkRequests` | 📋 Planned | - |
| `status` | `any` | `__vm.status` | 📋 Planned | - |
| `syntheticModule.setExport(name, value)` | `(...) => any` | `__vm.syntheticModule.setExport` | 📋 Planned | - |
| `vm.Module` | `(...) => any` | `__vm.vm.Module` | 📋 Planned | - |
| `vm.Script` | `(...) => any` | `__vm.vm.Script` | 📋 Planned | - |
| `vm.SourceTextModule` | `(...) => any` | `__vm.vm.SourceTextModule` | 📋 Planned | - |
| `vm.SyntheticModule` | `(...) => any` | `__vm.vm.SyntheticModule` | 📋 Planned | - |
| `vm.compileFunction(code[, params[, options]])` | `(...) => any` | `__vm.vm.compileFunction` | 📋 Planned | - |
| `vm.constants.DONT_CONTEXTIFY` | `any` | `__vm.vm.constants.DONT_CONTEXTIFY` | 📋 Planned | - |
| `vm.createContext([contextObject[, options]])` | `(...) => any` | `__vm.vm.createContext` | 📋 Planned | - |
| `vm.isContext(object)` | `(...) => any` | `__vm.vm.isContext` | 📋 Planned | - |
| `vm.measureMemory([options])` | `(...) => any` | `__vm.vm.measureMemory` | 📋 Planned | - |
| `vm.runInContext(code, contextifiedObject[, options])` | `(...) => any` | `__vm.vm.runInContext` | 📋 Planned | - |
| `vm.runInNewContext(code[, contextObject[, options]])` | `(...) => any` | `__vm.vm.runInNewContext` | 📋 Planned | - |
| `vm.runInThisContext(code[, options])` | `(...) => any` | `__vm.vm.runInThisContext` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `vm` are organized per API under `internal/compiler/testdata/corpus/vm/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/vm/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
