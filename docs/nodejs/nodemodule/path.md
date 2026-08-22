# Path Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:path`  
> **Specification Reference**: [Node.js 22 LTS Path Documentation](https://nodejs.org/docs/latest-v22.x/api/path.html)  
> **Type Definition Source**: [@types/node/path.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-path-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:path`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `path.basename(path[, suffix])` | `(...) => any` | `__path.path.basename` | ✅ Done | `internal/compiler/testdata/corpus/api/path.ts` |
| `path.dirname(path)` | `(...) => any` | `__path.path.dirname` | ✅ Done | `internal/compiler/testdata/corpus/api/path.ts` |
| `path.extname(path)` | `(...) => any` | `__path.path.extname` | ✅ Done | `internal/compiler/testdata/corpus/api/path.ts` |
| `path.join([...paths])` | `(...) => any` | `__path.path.join` | ✅ Done | `internal/compiler/testdata/corpus/api/path.ts` |
| `delimiter` | `any` | `__path.delimiter` | 📋 Planned | - |
| `path.format(pathObject)` | `(...) => any` | `__path.path.format` | 📋 Planned | - |
| `path.isAbsolute(path)` | `(...) => any` | `__path.path.isAbsolute` | 📋 Planned | - |
| `path.matchesGlob(path, pattern)` | `(...) => any` | `__path.path.matchesGlob` | 📋 Planned | - |
| `path.normalize(path)` | `(...) => any` | `__path.path.normalize` | 📋 Planned | - |
| `path.parse(path)` | `(...) => any` | `__path.path.parse` | 📋 Planned | - |
| `path.relative(from, to)` | `(...) => any` | `__path.path.relative` | 📋 Planned | - |
| `path.resolve([...paths])` | `(...) => any` | `__path.path.resolve` | 📋 Planned | - |
| `path.toNamespacedPath(path)` | `(...) => any` | `__path.path.toNamespacedPath` | 📋 Planned | - |
| `posix` | `any` | `__path.posix` | 📋 Planned | - |
| `sep` | `any` | `__path.sep` | 📋 Planned | - |
| `win32` | `any` | `__path.win32` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `path` are organized per API under `internal/compiler/testdata/corpus/path/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/path/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
