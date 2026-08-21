# Test runner Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:test_runner`  
> **Specification Reference**: [Node.js 22 LTS Test runner Documentation](https://nodejs.org/docs/latest-v22.x/api/test_runner.html)  
> **Type Definition Source**: [@types/node/test_runner.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-test_runner-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:test_runner`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `MockFunctionContext` | `(...) => any` | `__test_runner.MockFunctionContext` | 📋 Planned | - |
| `MockModuleContext` | `(...) => any` | `__test_runner.MockModuleContext` | 📋 Planned | - |
| `MockPropertyContext` | `(...) => any` | `__test_runner.MockPropertyContext` | 📋 Planned | - |
| `MockTimers` | `(...) => any` | `__test_runner.MockTimers` | 📋 Planned | - |
| `MockTracker` | `(...) => any` | `__test_runner.MockTracker` | 📋 Planned | - |
| `SuiteContext` | `(...) => any` | `__test_runner.SuiteContext` | 📋 Planned | - |
| `TestContext` | `(...) => any` | `__test_runner.TestContext` | 📋 Planned | - |
| `TestsStream` | `(...) => any` | `__test_runner.TestsStream` | 📋 Planned | - |
| `accesses` {Array}` | `any` | `__test_runner.accesses` {Array}` | 📋 Planned | - |
| `after([fn][, options])` | `(...) => any` | `__test_runner.after` | 📋 Planned | - |
| `afterEach([fn][, options])` | `(...) => any` | `__test_runner.afterEach` | 📋 Planned | - |
| `assert.register(name, fn)` | `(...) => any` | `__test_runner.assert.register` | 📋 Planned | - |
| `attempt` | `any` | `__test_runner.attempt` | 📋 Planned | - |
| `before([fn][, options])` | `(...) => any` | `__test_runner.before` | 📋 Planned | - |
| `beforeEach([fn][, options])` | `(...) => any` | `__test_runner.beforeEach` | 📋 Planned | - |
| `calls` | `any` | `__test_runner.calls` | 📋 Planned | - |
| `context.after([fn][, options])` | `(...) => any` | `__test_runner.context.after` | 📋 Planned | - |
| `context.afterEach([fn][, options])` | `(...) => any` | `__test_runner.context.afterEach` | 📋 Planned | - |
| `context.assert` | `any` | `__test_runner.context.assert` | 📋 Planned | - |
| `context.before([fn][, options])` | `(...) => any` | `__test_runner.context.before` | 📋 Planned | - |
| `context.beforeEach([fn][, options])` | `(...) => any` | `__test_runner.context.beforeEach` | 📋 Planned | - |
| `context.diagnostic(message)` | `(...) => any` | `__test_runner.context.diagnostic` | 📋 Planned | - |
| `context.filePath` | `any` | `__test_runner.context.filePath` | 📋 Planned | - |
| `context.fullName` | `any` | `__test_runner.context.fullName` | 📋 Planned | - |
| `context.name` | `any` | `__test_runner.context.name` | 📋 Planned | - |
| `context.plan(count[,options])` | `(...) => any` | `__test_runner.context.plan` | 📋 Planned | - |
| `context.runOnly(shouldRunOnlyTests)` | `(...) => any` | `__test_runner.context.runOnly` | 📋 Planned | - |
| `context.skip([message])` | `(...) => any` | `__test_runner.context.skip` | 📋 Planned | - |
| `context.test([name][, options][, fn])` | `(...) => any` | `__test_runner.context.test` | 📋 Planned | - |
| `context.todo([message])` | `(...) => any` | `__test_runner.context.todo` | 📋 Planned | - |
| `context.waitFor(condition[, options])` | `(...) => any` | `__test_runner.context.waitFor` | 📋 Planned | - |
| `ctx.accessCount()` | `(...) => any` | `__test_runner.ctx.accessCount` | 📋 Planned | - |
| `ctx.callCount()` | `(...) => any` | `__test_runner.ctx.callCount` | 📋 Planned | - |
| `ctx.mockImplementation(implementation)` | `(...) => any` | `__test_runner.ctx.mockImplementation` | 📋 Planned | - |
| `ctx.mockImplementation(value)` | `(...) => any` | `__test_runner.ctx.mockImplementation` | 📋 Planned | - |
| `ctx.mockImplementationOnce(implementation[, onCall])` | `(...) => any` | `__test_runner.ctx.mockImplementationOnce` | 📋 Planned | - |
| `ctx.mockImplementationOnce(value[, onAccess])` | `(...) => any` | `__test_runner.ctx.mockImplementationOnce` | 📋 Planned | - |
| `ctx.resetAccesses()` | `(...) => any` | `__test_runner.ctx.resetAccesses` | 📋 Planned | - |
| `ctx.resetCalls()` | `(...) => any` | `__test_runner.ctx.resetCalls` | 📋 Planned | - |
| `ctx.restore()` | `(...) => any` | `__test_runner.ctx.restore` | 📋 Planned | - |
| `describe([name][, options][, fn])` | `(...) => any` | `__test_runner.describe` | 📋 Planned | - |
| `describe.only([name][, options][, fn])` | `(...) => any` | `__test_runner.describe.only` | 📋 Planned | - |
| `describe.skip([name][, options][, fn])` | `(...) => any` | `__test_runner.describe.skip` | 📋 Planned | - |
| `describe.todo([name][, options][, fn])` | `(...) => any` | `__test_runner.describe.todo` | 📋 Planned | - |
| `error` | `any` | `__test_runner.error` | 📋 Planned | - |
| `it([name][, options][, fn])` | `(...) => any` | `__test_runner.it` | 📋 Planned | - |
| `it.only([name][, options][, fn])` | `(...) => any` | `__test_runner.it.only` | 📋 Planned | - |
| `it.skip([name][, options][, fn])` | `(...) => any` | `__test_runner.it.skip` | 📋 Planned | - |
| `it.todo([name][, options][, fn])` | `(...) => any` | `__test_runner.it.todo` | 📋 Planned | - |
| `mock.fn([original[, implementation]][, options])` | `(...) => any` | `__test_runner.mock.fn` | 📋 Planned | - |
| `mock.getter(object, methodName[, implementation][, options])` | `(...) => any` | `__test_runner.mock.getter` | 📋 Planned | - |
| `mock.method(object, methodName[, implementation][, options])` | `(...) => any` | `__test_runner.mock.method` | 📋 Planned | - |
| `mock.module(specifier[, options])` | `(...) => any` | `__test_runner.mock.module` | 📋 Planned | - |
| `mock.property(object, propertyName[, value])` | `(...) => any` | `__test_runner.mock.property` | 📋 Planned | - |
| `mock.reset()` | `(...) => any` | `__test_runner.mock.reset` | 📋 Planned | - |
| `mock.restoreAll()` | `(...) => any` | `__test_runner.mock.restoreAll` | 📋 Planned | - |
| `mock.setter(object, methodName[, implementation][, options])` | `(...) => any` | `__test_runner.mock.setter` | 📋 Planned | - |
| `passed` | `any` | `__test_runner.passed` | 📋 Planned | - |
| `run([options])` | `(...) => any` | `__test_runner.run` | 📋 Planned | - |
| `signal` | `any` | `__test_runner.signal` | 📋 Planned | - |
| `snapshot.setDefaultSnapshotSerializers(serializers)` | `(...) => any` | `__test_runner.snapshot.setDefaultSnapshotSerializers` | 📋 Planned | - |
| `snapshot.setResolveSnapshotPath(fn)` | `(...) => any` | `__test_runner.snapshot.setResolveSnapshotPath` | 📋 Planned | - |
| `suite([name][, options][, fn])` | `(...) => any` | `__test_runner.suite` | 📋 Planned | - |
| `suite.only([name][, options][, fn])` | `(...) => any` | `__test_runner.suite.only` | 📋 Planned | - |
| `suite.skip([name][, options][, fn])` | `(...) => any` | `__test_runner.suite.skip` | 📋 Planned | - |
| `suite.todo([name][, options][, fn])` | `(...) => any` | `__test_runner.suite.todo` | 📋 Planned | - |
| `test([name][, options][, fn])` | `(...) => any` | `__test_runner.test` | 📋 Planned | - |
| `test.only([name][, options][, fn])` | `(...) => any` | `__test_runner.test.only` | 📋 Planned | - |
| `test.skip([name][, options][, fn])` | `(...) => any` | `__test_runner.test.skip` | 📋 Planned | - |
| `test.todo([name][, options][, fn])` | `(...) => any` | `__test_runner.test.todo` | 📋 Planned | - |
| `timers.enable([enableOptions])` | `(...) => any` | `__test_runner.timers.enable` | 📋 Planned | - |
| `timers.reset()` | `(...) => any` | `__test_runner.timers.reset` | 📋 Planned | - |
| `timers.runAll()` | `(...) => any` | `__test_runner.timers.runAll` | 📋 Planned | - |
| `timers.setTime(milliseconds)` | `(...) => any` | `__test_runner.timers.setTime` | 📋 Planned | - |
| `timers.tick([milliseconds])` | `(...) => any` | `__test_runner.timers.tick` | 📋 Planned | - |
| `timers[Symbol.dispose]()` | `(...) => any` | `__test_runner.timers[Symbol.dispose]` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `test_runner` are organized per API under `internal/compiler/testdata/corpus/test_runner/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/test_runner/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
