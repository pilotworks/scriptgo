# Child process Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:child_process`  
> **Specification Reference**: [Node.js 22 LTS Child process Documentation](https://nodejs.org/docs/latest-v22.x/api/child_process.html)  
> **Type Definition Source**: [@types/node/child_process.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-child_process-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:child_process`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `child_process.execSync(command[, options])` | `(...) => any` | `__child_process.child_process.execSync` | ✅ Done | `internal/compiler/testdata/corpus/api/child_process.ts` |
| `child_process.spawnSync(command[, args][, options])` | `(...) => any` | `__child_process.child_process.spawnSync` | ✅ Done | `internal/compiler/testdata/corpus/api/child_process.ts` |
| `ChildProcess` | `(...) => any` | `__child_process.ChildProcess` | 📋 Planned | - |
| `channel` | `any` | `__child_process.channel` | 📋 Planned | - |
| `child_process.exec(command[, options][, callback])` | `(...) => any` | `__child_process.child_process.exec` | 📋 Planned | - |
| `child_process.execFile(file[, args][, options][, callback])` | `(...) => any` | `__child_process.child_process.execFile` | 📋 Planned | - |
| `child_process.execFileSync(file[, args][, options])` | `(...) => any` | `__child_process.child_process.execFileSync` | 📋 Planned | - |
| `child_process.fork(modulePath[, args][, options])` | `(...) => any` | `__child_process.child_process.fork` | 📋 Planned | - |
| `child_process.spawn(command[, args][, options])` | `(...) => any` | `__child_process.child_process.spawn` | 📋 Planned | - |
| `connected` | `any` | `__child_process.connected` | 📋 Planned | - |
| `exitCode` | `any` | `__child_process.exitCode` | 📋 Planned | - |
| `killed` | `any` | `__child_process.killed` | 📋 Planned | - |
| `pid` | `any` | `__child_process.pid` | 📋 Planned | - |
| `signalCode` | `any` | `__child_process.signalCode` | 📋 Planned | - |
| `spawnargs` | `any` | `__child_process.spawnargs` | 📋 Planned | - |
| `spawnfile` | `any` | `__child_process.spawnfile` | 📋 Planned | - |
| `stderr` | `any` | `__child_process.stderr` | 📋 Planned | - |
| `stdin` | `any` | `__child_process.stdin` | 📋 Planned | - |
| `stdio` | `any` | `__child_process.stdio` | 📋 Planned | - |
| `stdout` | `any` | `__child_process.stdout` | 📋 Planned | - |
| `subprocess.disconnect()` | `(...) => any` | `__child_process.subprocess.disconnect` | 📋 Planned | - |
| `subprocess.kill([signal])` | `(...) => any` | `__child_process.subprocess.kill` | 📋 Planned | - |
| `subprocess.ref()` | `(...) => any` | `__child_process.subprocess.ref` | 📋 Planned | - |
| `subprocess.send(message[, sendHandle[, options]][, callback])` | `(...) => any` | `__child_process.subprocess.send` | 📋 Planned | - |
| `subprocess.unref()` | `(...) => any` | `__child_process.subprocess.unref` | 📋 Planned | - |
| `subprocess[Symbol.dispose]()` | `(...) => any` | `__child_process.subprocess[Symbol.dispose]` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `child_process` are organized per API under `internal/compiler/testdata/corpus/child_process/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/child_process/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
