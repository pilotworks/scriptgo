# Event & EventTarget API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG Event & EventTarget API Specification](https://wintercg.org/)  
> **Type Definition Source**: [microsoft/TypeScript lib.dom.d.ts (Server subset)](https://github.com/microsoft/TypeScript/blob/main/src/lib/lib.dom.d.ts)  
> **Gate Oracle**: Web Platform Tests (WPT) & Node.js 22 LTS WPT runner

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
| `CustomEvent` | `(...) => any` | `__events.CustomEvent` | 📋 Planned | - |
| `Event` | `(...) => any` | `__events.Event` | 📋 Planned | - |
| `EventEmitter` | `(...) => any` | `__events.EventEmitter` | 📋 Planned | - |
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
| `emitter.emit(eventName[, ...args])` | `(...) => any` | `__events.emitter.emit` | 📋 Planned | - |
| `emitter.eventNames()` | `(...) => any` | `__events.emitter.eventNames` | 📋 Planned | - |
| `emitter.getMaxListeners()` | `(...) => any` | `__events.emitter.getMaxListeners` | 📋 Planned | - |
| `emitter.listenerCount(eventName[, listener])` | `(...) => any` | `__events.emitter.listenerCount` | 📋 Planned | - |
| `emitter.listeners(eventName)` | `(...) => any` | `__events.emitter.listeners` | 📋 Planned | - |
| `emitter.off(eventName, listener)` | `(...) => any` | `__events.emitter.off` | 📋 Planned | - |
| `emitter.on(eventName, listener)` | `(...) => any` | `__events.emitter.on` | 📋 Planned | - |
| `emitter.once(eventName, listener)` | `(...) => any` | `__events.emitter.once` | 📋 Planned | - |
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
| `events.listenerCount(emitter, eventName)` | `(...) => any` | `__events.events.listenerCount` | 📋 Planned | - |
| `events.on(emitter, eventName[, options])` | `(...) => any` | `__events.events.on` | 📋 Planned | - |
| `events.once(emitter, name[, options])` | `(...) => any` | `__events.events.once` | 📋 Planned | - |
| `events.setMaxListeners(n[, ...eventTargets])` | `(...) => any` | `__events.events.setMaxListeners` | 📋 Planned | - |
| `isTrusted` | `any` | `__events.isTrusted` | 📋 Planned | - |
| `nodeEventTarget.addListener(type, listener)` | `(...) => any` | `__events.nodeEventTarget.addListener` | 📋 Planned | - |
| `nodeEventTarget.emit(type, arg)` | `(...) => any` | `__events.nodeEventTarget.emit` | 📋 Planned | - |
| `nodeEventTarget.eventNames()` | `(...) => any` | `__events.nodeEventTarget.eventNames` | 📋 Planned | - |
| `nodeEventTarget.getMaxListeners()` | `(...) => any` | `__events.nodeEventTarget.getMaxListeners` | 📋 Planned | - |
| `nodeEventTarget.listenerCount(type)` | `(...) => any` | `__events.nodeEventTarget.listenerCount` | 📋 Planned | - |
| `nodeEventTarget.off(type, listener[, options])` | `(...) => any` | `__events.nodeEventTarget.off` | 📋 Planned | - |
| `nodeEventTarget.on(type, listener)` | `(...) => any` | `__events.nodeEventTarget.on` | 📋 Planned | - |
| `nodeEventTarget.once(type, listener)` | `(...) => any` | `__events.nodeEventTarget.once` | 📋 Planned | - |
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
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/events/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
