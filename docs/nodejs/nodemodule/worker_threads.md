# Worker threads Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:worker_threads`  
> **Specification Reference**: [Node.js 22 LTS Worker threads Documentation](https://nodejs.org/docs/latest-v22.x/api/worker_threads.html)  
> **Type Definition Source**: [@types/node/worker_threads.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-worker_threads-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:worker_threads`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `MessageChannel` | `(...) => any` | `__worker_threads.MessageChannel` | 📋 Planned | - |
| `MessagePort` | `(...) => any` | `__worker_threads.MessagePort` | 📋 Planned | - |
| `SHARE_ENV` | `any` | `__worker_threads.SHARE_ENV` | 📋 Planned | - |
| `Worker` | `(...) => any` | `__worker_threads.Worker` | 📋 Planned | - |
| `broadcastChannel.close()` | `(...) => any` | `__worker_threads.broadcastChannel.close` | 📋 Planned | - |
| `broadcastChannel.postMessage(message)` | `(...) => any` | `__worker_threads.broadcastChannel.postMessage` | 📋 Planned | - |
| `broadcastChannel.ref()` | `(...) => any` | `__worker_threads.broadcastChannel.ref` | 📋 Planned | - |
| `broadcastChannel.unref()` | `(...) => any` | `__worker_threads.broadcastChannel.unref` | 📋 Planned | - |
| `isInternalThread` | `any` | `__worker_threads.isInternalThread` | 📋 Planned | - |
| `isMainThread` | `any` | `__worker_threads.isMainThread` | 📋 Planned | - |
| `onmessage` | `any` | `__worker_threads.onmessage` | 📋 Planned | - |
| `onmessageerror` | `any` | `__worker_threads.onmessageerror` | 📋 Planned | - |
| `parentPort` | `any` | `__worker_threads.parentPort` | 📋 Planned | - |
| `port.close()` | `(...) => any` | `__worker_threads.port.close` | 📋 Planned | - |
| `port.hasRef()` | `(...) => any` | `__worker_threads.port.hasRef` | 📋 Planned | - |
| `port.postMessage(value[, transferList])` | `(...) => any` | `__worker_threads.port.postMessage` | 📋 Planned | - |
| `port.ref()` | `(...) => any` | `__worker_threads.port.ref` | 📋 Planned | - |
| `port.start()` | `(...) => any` | `__worker_threads.port.start` | 📋 Planned | - |
| `port.unref()` | `(...) => any` | `__worker_threads.port.unref` | 📋 Planned | - |
| `resourceLimits` | `any` | `__worker_threads.resourceLimits` | 📋 Planned | - |
| `stderr` | `any` | `__worker_threads.stderr` | 📋 Planned | - |
| `stdin` | `any` | `__worker_threads.stdin` | 📋 Planned | - |
| `stdout` | `any` | `__worker_threads.stdout` | 📋 Planned | - |
| `threadId` | `any` | `__worker_threads.threadId` | 📋 Planned | - |
| `threadName` {string\|null}` | `any` | `__worker_threads.threadName` {string\|null}` | 📋 Planned | - |
| `worker.cpuUsage([prev])` | `(...) => any` | `__worker_threads.worker.cpuUsage` | 📋 Planned | - |
| `worker.getEnvironmentData(key)` | `(...) => any` | `__worker_threads.worker.getEnvironmentData` | 📋 Planned | - |
| `worker.getHeapSnapshot([options])` | `(...) => any` | `__worker_threads.worker.getHeapSnapshot` | 📋 Planned | - |
| `worker.getHeapStatistics()` | `(...) => any` | `__worker_threads.worker.getHeapStatistics` | 📋 Planned | - |
| `worker.isMarkedAsUntransferable(object)` | `(...) => any` | `__worker_threads.worker.isMarkedAsUntransferable` | 📋 Planned | - |
| `worker.markAsUncloneable(object)` | `(...) => any` | `__worker_threads.worker.markAsUncloneable` | 📋 Planned | - |
| `worker.markAsUntransferable(object)` | `(...) => any` | `__worker_threads.worker.markAsUntransferable` | 📋 Planned | - |
| `worker.moveMessagePortToContext(port, contextifiedSandbox)` | `(...) => any` | `__worker_threads.worker.moveMessagePortToContext` | 📋 Planned | - |
| `worker.performance` | `any` | `__worker_threads.worker.performance` | 📋 Planned | - |
| `worker.postMessage(value[, transferList])` | `(...) => any` | `__worker_threads.worker.postMessage` | 📋 Planned | - |
| `worker.postMessageToThread(threadId, value[, transferList][, timeout])` | `(...) => any` | `__worker_threads.worker.postMessageToThread` | 📋 Planned | - |
| `worker.receiveMessageOnPort(port)` | `(...) => any` | `__worker_threads.worker.receiveMessageOnPort` | 📋 Planned | - |
| `worker.ref()` | `(...) => any` | `__worker_threads.worker.ref` | 📋 Planned | - |
| `worker.setEnvironmentData(key[, value])` | `(...) => any` | `__worker_threads.worker.setEnvironmentData` | 📋 Planned | - |
| `worker.startCpuProfile(name)` | `(...) => any` | `__worker_threads.worker.startCpuProfile` | 📋 Planned | - |
| `worker.terminate()` | `(...) => any` | `__worker_threads.worker.terminate` | 📋 Planned | - |
| `worker.unref()` | `(...) => any` | `__worker_threads.worker.unref` | 📋 Planned | - |
| `worker.workerData` | `any` | `__worker_threads.worker.workerData` | 📋 Planned | - |
| `worker[Symbol.asyncDispose]()` | `(...) => any` | `__worker_threads.worker[Symbol.asyncDispose]` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `worker_threads` are organized per API under `internal/compiler/testdata/corpus/worker_threads/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/worker_threads/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
