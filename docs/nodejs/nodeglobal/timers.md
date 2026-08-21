# Timers Global Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS Timers Global Documentation](https://nodejs.org/docs/latest-v22.x/api/timers.html)  
> **Type Definition Source**: [@types/node/timers.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-global-*.js, test-timers-*.js)

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
| `Timeout` | `(...) => any` | `__timers.Timeout` | ✅ Done | `internal/compiler/testdata/corpus/timers/timeout/` |
| `Immediate` | `(...) => any` | `__timers.Immediate` | 📋 Planned | - |
| `clearImmediate(immediate)` | `(...) => any` | `__timers.clearImmediate` | 📋 Planned | - |
| `clearInterval(timeout)` | `(...) => any` | `__timers.clearInterval` | 📋 Planned | - |
| `clearTimeout(timeout)` | `(...) => any` | `__timers.clearTimeout` | 📋 Planned | - |
| `immediate.hasRef()` | `(...) => any` | `__timers.immediate.hasRef` | 📋 Planned | - |
| `immediate.ref()` | `(...) => any` | `__timers.immediate.ref` | 📋 Planned | - |
| `immediate.unref()` | `(...) => any` | `__timers.immediate.unref` | 📋 Planned | - |
| `immediate[Symbol.dispose]()` | `(...) => any` | `__timers.immediate[Symbol.dispose]` | 📋 Planned | - |
| `setImmediate(callback[, ...args])` | `(...) => any` | `__timers.setImmediate` | 📋 Planned | - |
| `setInterval(callback[, delay[, ...args]])` | `(...) => any` | `__timers.setInterval` | 📋 Planned | - |
| `setTimeout(callback[, delay[, ...args]])` | `(...) => any` | `__timers.setTimeout` | 📋 Planned | - |
| `timeout.close()` | `(...) => any` | `__timers.timeout.close` | 📋 Planned | - |
| `timeout.hasRef()` | `(...) => any` | `__timers.timeout.hasRef` | 📋 Planned | - |
| `timeout.ref()` | `(...) => any` | `__timers.timeout.ref` | 📋 Planned | - |
| `timeout.refresh()` | `(...) => any` | `__timers.timeout.refresh` | 📋 Planned | - |
| `timeout.unref()` | `(...) => any` | `__timers.timeout.unref` | 📋 Planned | - |
| `timeout[Symbol.dispose]()` | `(...) => any` | `__timers.timeout[Symbol.dispose]` | 📋 Planned | - |
| `timeout[Symbol.toPrimitive]()` | `(...) => any` | `__timers.timeout[Symbol.toPrimitive]` | 📋 Planned | - |
| `timersPromises.scheduler.wait(delay[, options])` | `(...) => any` | `__timers.timersPromises.scheduler.wait` | 📋 Planned | - |
| `timersPromises.scheduler.yield()` | `(...) => any` | `__timers.timersPromises.scheduler.yield` | 📋 Planned | - |
| `timersPromises.setImmediate([value[, options]])` | `(...) => any` | `__timers.timersPromises.setImmediate` | 📋 Planned | - |
| `timersPromises.setInterval([delay[, value[, options]]])` | `(...) => any` | `__timers.timersPromises.setInterval` | 📋 Planned | - |
| `timersPromises.setTimeout([delay[, value[, options]]])` | `(...) => any` | `__timers.timersPromises.setTimeout` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `timers` are organized per API under `internal/compiler/testdata/corpus/timers/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/timers/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
