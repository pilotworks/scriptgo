# Modules: `node:module` API Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:module`  
> **Specification Reference**: [Node.js 22 LTS Modules: `node:module` API Documentation](https://nodejs.org/docs/latest-v22.x/api/module.html)  
> **Type Definition Source**: [@types/node/module.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-module-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:module`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `builtinModules` | `any` | `__module.builtinModules` | 📋 Planned | - |
| `module.SourceMap` | `(...) => any` | `__module.module.SourceMap` | 📋 Planned | - |
| `module.constants.compileCacheStatus` | `any` | `__module.module.constants.compileCacheStatus` | 📋 Planned | - |
| `module.createRequire(filename)` | `(...) => any` | `__module.module.createRequire` | 📋 Planned | - |
| `module.enableCompileCache([cacheDir])` | `(...) => any` | `__module.module.enableCompileCache` | 📋 Planned | - |
| `module.findPackageJSON(specifier[, base])` | `(...) => any` | `__module.module.findPackageJSON` | 📋 Planned | - |
| `module.findSourceMap(path)` | `(...) => any` | `__module.module.findSourceMap` | 📋 Planned | - |
| `module.flushCompileCache()` | `(...) => any` | `__module.module.flushCompileCache` | 📋 Planned | - |
| `module.getCompileCacheDir()` | `(...) => any` | `__module.module.getCompileCacheDir` | 📋 Planned | - |
| `module.getSourceMapsSupport()` | `(...) => any` | `__module.module.getSourceMapsSupport` | 📋 Planned | - |
| `module.isBuiltin(moduleName)` | `(...) => any` | `__module.module.isBuiltin` | 📋 Planned | - |
| `module.register(specifier[, parentURL][, options])` | `(...) => any` | `__module.module.register` | 📋 Planned | - |
| `module.registerHooks(options)` | `(...) => any` | `__module.module.registerHooks` | 📋 Planned | - |
| `module.setSourceMapsSupport(enabled[, options])` | `(...) => any` | `__module.module.setSourceMapsSupport` | 📋 Planned | - |
| `module.stripTypeScriptTypes(code[, options])` | `(...) => any` | `__module.module.stripTypeScriptTypes` | 📋 Planned | - |
| `module.syncBuiltinESMExports()` | `(...) => any` | `__module.module.syncBuiltinESMExports` | 📋 Planned | - |
| `payload` Returns: {Object}` | `any` | `__module.payload` Returns: {Object}` | 📋 Planned | - |
| `sourceMap.findEntry(lineOffset, columnOffset)` | `(...) => any` | `__module.sourceMap.findEntry` | 📋 Planned | - |
| `sourceMap.findOrigin(lineNumber, columnNumber)` | `(...) => any` | `__module.sourceMap.findOrigin` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `module` are organized per API under `internal/compiler/testdata/corpus/module/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/module/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
