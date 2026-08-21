# Console Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Console Documentation](https://nodejs.org/docs/latest-v22.x/api/console.html)  
> **Type Definition Source**: [@types/node/console.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-console-*.js)

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
| `console.assert(value[, ...message])` | `(...) => any` | `__console.console.assert` | ✅ Done | `internal/compiler/testdata/corpus/api/console/assert/` |
| `console.clear()` | `(...) => any` | `__console.console.clear` | ✅ Done | `internal/compiler/testdata/corpus/api/console/clear/` |
| `console.count([label])` | `(...) => any` | `__console.console.count` | ✅ Done | `internal/compiler/testdata/corpus/api/console/count/` |
| `console.debug(data[, ...args])` | `(...) => any` | `__console.console.debug` | ✅ Done | `internal/compiler/testdata/corpus/api/console/debug/` |
| `console.dir(obj[, options])` | `(...) => any` | `__console.console.dir` | ✅ Done | `internal/compiler/testdata/corpus/api/console/dir/` |
| `console.error([data][, ...args])` | `(...) => any` | `__console.console.error` | ✅ Done | `internal/compiler/testdata/corpus/api/console/error/` |
| `console.group([...label])` | `(...) => any` | `__console.console.group` | ✅ Done | `internal/compiler/testdata/corpus/api/console/group/` |
| `console.groupEnd()` | `(...) => any` | `__console.console.groupEnd` | ✅ Done | `internal/compiler/testdata/corpus/api/console/groupEnd/` |
| `console.info([data][, ...args])` | `(...) => any` | `__console.console.info` | ✅ Done | `internal/compiler/testdata/corpus/api/console/info/` |
| `console.log([data][, ...args])` | `(...) => any` | `__console.console.log` | ✅ Done | `internal/compiler/testdata/corpus/api/console/log/` |
| `console.time([label])` | `(...) => any` | `__console.console.time` | ✅ Done | `internal/compiler/testdata/corpus/api/console/time/` |
| `console.trace([message][, ...args])` | `(...) => any` | `__console.console.trace` | ✅ Done | `internal/compiler/testdata/corpus/api/console/trace/` |
| `console.warn([data][, ...args])` | `(...) => any` | `__console.console.warn` | ✅ Done | `internal/compiler/testdata/corpus/api/console/warn/` |
| `Console` | `(...) => any` | `__console.Console` | 📋 Planned | - |
| `console.countReset([label])` | `(...) => any` | `__console.console.countReset` | 📋 Planned | - |
| `console.dirxml(...data)` | `(...) => any` | `__console.console.dirxml` | 📋 Planned | - |
| `console.groupCollapsed()` | `(...) => any` | `__console.console.groupCollapsed` | 📋 Planned | - |
| `console.profile([label])` | `(...) => any` | `__console.console.profile` | 📋 Planned | - |
| `console.profileEnd([label])` | `(...) => any` | `__console.console.profileEnd` | 📋 Planned | - |
| `console.table(tabularData[, properties])` | `(...) => any` | `__console.console.table` | 📋 Planned | - |
| `console.timeEnd([label])` | `(...) => any` | `__console.console.timeEnd` | 📋 Planned | - |
| `console.timeLog([label][, ...data])` | `(...) => any` | `__console.console.timeLog` | 📋 Planned | - |
| `console.timeStamp([label])` | `(...) => any` | `__console.console.timeStamp` | 📋 Planned | - |

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
