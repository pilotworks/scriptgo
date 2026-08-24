# REPL Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:repl`  
> **Specification Reference**: [Node.js 22 LTS REPL Documentation](https://nodejs.org/docs/latest-v22.x/api/repl.html)  
> **Type Definition Source**: [@types/node/repl.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-repl-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:repl`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `REPLServer` | `(...) => any` | `__repl.REPLServer` | 📋 Planned | - |
| `builtinModules` | `any` | `__repl.builtinModules` | 📋 Planned | - |
| `repl.start([options])` | `(...) => any` | `__repl.repl.start` | 📋 Planned | - |
| `replServer.clearBufferedCommand()` | `(...) => any` | `__repl.replServer.clearBufferedCommand` | 📋 Planned | - |
| `replServer.defineCommand(keyword, cmd)` | `(...) => any` | `__repl.replServer.defineCommand` | 📋 Planned | - |
| `replServer.displayPrompt([preserveCursor])` | `(...) => any` | `__repl.replServer.displayPrompt` | 📋 Planned | - |
| `replServer.setupHistory(historyPath, callback)` | `(...) => any` | `__repl.replServer.setupHistory` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `repl` are organized per API under `internal/compiler/testdata/corpus/repl/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/repl/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
