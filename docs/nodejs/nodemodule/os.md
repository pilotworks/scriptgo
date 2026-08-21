# OS Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:os`  
> **Specification Reference**: [Node.js 22 LTS OS Documentation](https://nodejs.org/docs/latest-v22.x/api/os.html)  
> **Type Definition Source**: [@types/node/os.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-os-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:os`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `os.arch()` | `(...) => any` | `__os.os.arch` | ✅ Done | `internal/compiler/testdata/corpus/api/os/arch/` |
| `os.freemem()` | `(...) => any` | `__os.os.freemem` | ✅ Done | `internal/compiler/testdata/corpus/api/os/freemem/` |
| `os.homedir()` | `(...) => any` | `__os.os.homedir` | ✅ Done | `internal/compiler/testdata/corpus/api/os/homedir/` |
| `os.platform()` | `(...) => any` | `__os.os.platform` | ✅ Done | `internal/compiler/testdata/corpus/api/os/platform/` |
| `os.release()` | `(...) => any` | `__os.os.release` | ✅ Done | `internal/compiler/testdata/corpus/api/os/release/` |
| `os.tmpdir()` | `(...) => any` | `__os.os.tmpdir` | ✅ Done | `internal/compiler/testdata/corpus/api/os/tmpdir/` |
| `os.totalmem()` | `(...) => any` | `__os.os.totalmem` | ✅ Done | `internal/compiler/testdata/corpus/api/os/totalmem/` |
| `os.type()` | `(...) => any` | `__os.os.type` | ✅ Done | `internal/compiler/testdata/corpus/api/os/type/` |
| `os.uptime()` | `(...) => any` | `__os.os.uptime` | ✅ Done | `internal/compiler/testdata/corpus/api/os/uptime/` |
| `EOL` | `any` | `__os.EOL` | 📋 Planned | - |
| `constants` | `any` | `__os.constants` | 📋 Planned | - |
| `devNull` | `any` | `__os.devNull` | 📋 Planned | - |
| `os.availableParallelism()` | `(...) => any` | `__os.os.availableParallelism` | 📋 Planned | - |
| `os.cpus()` | `(...) => any` | `__os.os.cpus` | 📋 Planned | - |
| `os.endianness()` | `(...) => any` | `__os.os.endianness` | 📋 Planned | - |
| `os.getPriority([pid])` | `(...) => any` | `__os.os.getPriority` | 📋 Planned | - |
| `os.hostname()` | `(...) => any` | `__os.os.hostname` | 📋 Planned | - |
| `os.loadavg()` | `(...) => any` | `__os.os.loadavg` | 📋 Planned | - |
| `os.machine()` | `(...) => any` | `__os.os.machine` | 📋 Planned | - |
| `os.networkInterfaces()` | `(...) => any` | `__os.os.networkInterfaces` | 📋 Planned | - |
| `os.setPriority([pid, ]priority)` | `(...) => any` | `__os.os.setPriority` | 📋 Planned | - |
| `os.userInfo([options])` | `(...) => any` | `__os.os.userInfo` | 📋 Planned | - |
| `os.version()` | `(...) => any` | `__os.os.version` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `os` are organized per API under `internal/compiler/testdata/corpus/os/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/os/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
