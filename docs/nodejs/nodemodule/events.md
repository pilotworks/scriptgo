# Events Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:events`  
> **Specification Reference**: [Node.js 22 LTS Events Documentation](https://nodejs.org/docs/latest-v22.x/api/events.html)  
> **Type Definition Source**: [@types/node/events.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-events-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:events`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `EventEmitter` | `(...) => any` | `__events.EventEmitter` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.emit(eventName[, ...args])` | `(...) => any` | `__events.emitter.emit` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.eventNames()` | `(...) => any` | `__events.emitter.eventNames` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.listenerCount(eventName[, listener])` | `(...) => any` | `__events.emitter.listenerCount` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.off(eventName, listener)` | `(...) => any` | `__events.emitter.off` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.on(eventName, listener)` | `(...) => any` | `__events.emitter.on` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `emitter.once(eventName, listener)` | `(...) => any` | `__events.emitter.once` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `events.listenerCount(emitter, eventName)` | `(...) => any` | `__events.events.listenerCount` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `events.on(emitter, eventName[, options])` | `(...) => any` | `__events.events.on` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `events.once(emitter, name[, options])` | `(...) => any` | `__events.events.once` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.emit(type, arg)` | `(...) => any` | `__events.nodeEventTarget.emit` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.eventNames()` | `(...) => any` | `__events.nodeEventTarget.eventNames` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.listenerCount(type)` | `(...) => any` | `__events.nodeEventTarget.listenerCount` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.off(type, listener[, options])` | `(...) => any` | `__events.nodeEventTarget.off` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.on(type, listener)` | `(...) => any` | `__events.nodeEventTarget.on` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `nodeEventTarget.once(type, listener)` | `(...) => any` | `__events.nodeEventTarget.once` | ✅ Done | `internal/compiler/testdata/corpus/api/events.ts` |
| `CustomEvent` | `(...) => any` | `__events.CustomEvent` | 📋 Planned | - |
| `Event` | `(...) => any` | `__events.Event` | 📋 Planned | - |
| `EventTarget` | `(...) => any` | `__events.EventTarget` | 📋 Planned | - |
| `NodeEventTarget` | `(...) => any` | `__events.NodeEventTarget` | 📋 Planned | - |
| `asyncId` | `any` | `__events.asyncId` | 📋 Planned | - |
| `asyncResource` | `any` | `__events.asyncResource` | 📋 Planned | - |
| `bubbles` | `any` | `__events.bubbles` | 📋 Planned | - |
| `cancelBubble` | `any` | `__events.cancelBubble` | 📋 Planned | - |
| `cancelable` | `any` | `__events.cancelable` | 📋 Planned | - |
| `captureRejectionSymbol` | `any` | `__events.captureRejectionSymbol` | 📋 Planned | - |
| `captureRejections` | `any` | `__events.captureRejections` | 📋 Planned | - |
| `composed` | `any` | `__events.composed` | 📋 Planned | - |
| `currentTarget` | `any` | `__events.currentTarget` | 📋 Planned | - |
| `defaultPrevented` | `any` | `__events.defaultPrevented` | 📋 Planned | - |
| `detail` | `any` | `__events.detail` | 📋 Planned | - |
| `emitter.addListener(eventName, listener)` | `(...) => any` | `__events.emitter.addListener` | 📋 Planned | - |
| `emitter.getMaxListeners()` | `(...) => any` | `__events.emitter.getMaxListeners` | 📋 Planned | - |
| `emitter.listeners(eventName)` | `(...) => any` | `__events.emitter.listeners` | 📋 Planned | - |
| `emitter.prependListener(eventName, listener)` | `(...) => any` | `__events.emitter.prependListener` | 📋 Planned | - |
| `emitter.prependOnceListener(eventName, listener)` | `(...) => any` | `__events.emitter.prependOnceListener` | 📋 Planned | - |
| `emitter.rawListeners(eventName)` | `(...) => any` | `__events.emitter.rawListeners` | 📋 Planned | - |
| `emitter.removeAllListeners([eventName])` | `(...) => any` | `__events.emitter.removeAllListeners` | 📋 Planned | - |
| `emitter.removeListener(eventName, listener)` | `(...) => any` | `__events.emitter.removeListener` | 📋 Planned | - |
| `emitter.setMaxListeners(n)` | `(...) => any` | `__events.emitter.setMaxListeners` | 📋 Planned | - |
| `event.composedPath()` | `(...) => any` | `__events.event.composedPath` | 📋 Planned | - |
| `event.initEvent(type[, bubbles[, cancelable]])` | `(...) => any` | `__events.event.initEvent` | 📋 Planned | - |
| `event.preventDefault()` | `(...) => any` | `__events.event.preventDefault` | 📋 Planned | - |
| `event.stopImmediatePropagation()` | `(...) => any` | `__events.event.stopImmediatePropagation` | 📋 Planned | - |
| `event.stopPropagation()` | `(...) => any` | `__events.event.stopPropagation` | 📋 Planned | - |
| `eventPhase` | `any` | `__events.eventPhase` | 📋 Planned | - |
| `eventTarget.addEventListener(type, listener[, options])` | `(...) => any` | `__events.eventTarget.addEventListener` | 📋 Planned | - |
| `eventTarget.dispatchEvent(event)` | `(...) => any` | `__events.eventTarget.dispatchEvent` | 📋 Planned | - |
| `eventTarget.removeEventListener(type, listener[, options])` | `(...) => any` | `__events.eventTarget.removeEventListener` | 📋 Planned | - |
| `eventemitterasyncresource.emitDestroy()` | `(...) => any` | `__events.eventemitterasyncresource.emitDestroy` | 📋 Planned | - |
| `events.EventEmitterAsyncResource extends EventEmitter` | `(...) => any` | `__events.events.EventEmitterAsyncResource extends EventEmitter` | 📋 Planned | - |
| `events.addAbortListener(signal, listener)` | `(...) => any` | `__events.events.addAbortListener` | 📋 Planned | - |
| `events.defaultMaxListeners` | `any` | `__events.events.defaultMaxListeners` | 📋 Planned | - |
| `events.errorMonitor` | `any` | `__events.events.errorMonitor` | 📋 Planned | - |
| `events.getEventListeners(emitterOrTarget, eventName)` | `(...) => any` | `__events.events.getEventListeners` | 📋 Planned | - |
| `events.getMaxListeners(emitterOrTarget)` | `(...) => any` | `__events.events.getMaxListeners` | 📋 Planned | - |
| `events.setMaxListeners(n[, ...eventTargets])` | `(...) => any` | `__events.events.setMaxListeners` | 📋 Planned | - |
| `isTrusted` | `any` | `__events.isTrusted` | 📋 Planned | - |
| `nodeEventTarget.addListener(type, listener)` | `(...) => any` | `__events.nodeEventTarget.addListener` | 📋 Planned | - |
| `nodeEventTarget.getMaxListeners()` | `(...) => any` | `__events.nodeEventTarget.getMaxListeners` | 📋 Planned | - |
| `nodeEventTarget.removeAllListeners([type])` | `(...) => any` | `__events.nodeEventTarget.removeAllListeners` | 📋 Planned | - |
| `nodeEventTarget.removeListener(type, listener[, options])` | `(...) => any` | `__events.nodeEventTarget.removeListener` | 📋 Planned | - |
| `nodeEventTarget.setMaxListeners(n)` | `(...) => any` | `__events.nodeEventTarget.setMaxListeners` | 📋 Planned | - |
| `returnValue` | `any` | `__events.returnValue` | 📋 Planned | - |
| `srcElement` | `any` | `__events.srcElement` | 📋 Planned | - |
| `target` | `any` | `__events.target` | 📋 Planned | - |
| `timeStamp` | `any` | `__events.timeStamp` | 📋 Planned | - |
| `triggerAsyncId` | `any` | `__events.triggerAsyncId` | 📋 Planned | - |
| `type` | `any` | `__events.type` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `events` are organized per API under `internal/compiler/testdata/corpus/events/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/events/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
