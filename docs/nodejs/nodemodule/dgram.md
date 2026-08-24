# dgram Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:dgram`  
> **Specification Reference**: [Node.js 22 LTS dgram Documentation](https://nodejs.org/docs/latest-v22.x/api/dgram.html)  
> **Type Definition Source**: [@types/node/dgram.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-dgram-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:dgram`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `dgram.Socket` | `(...) => any` | `__dgram.dgram.Socket` | 📋 Planned | - |
| `dgram.createSocket(options[, callback])` | `(...) => any` | `__dgram.dgram.createSocket` | 📋 Planned | - |
| `dgram.createSocket(type[, callback])` | `(...) => any` | `__dgram.dgram.createSocket` | 📋 Planned | - |
| `socket.addMembership(multicastAddress[, multicastInterface])` | `(...) => any` | `__dgram.socket.addMembership` | 📋 Planned | - |
| `socket.addSourceSpecificMembership(sourceAddress, groupAddress[, multicastInterface])` | `(...) => any` | `__dgram.socket.addSourceSpecificMembership` | 📋 Planned | - |
| `socket.address()` | `(...) => any` | `__dgram.socket.address` | 📋 Planned | - |
| `socket.bind([port][, address][, callback])` | `(...) => any` | `__dgram.socket.bind` | 📋 Planned | - |
| `socket.bind(options[, callback])` | `(...) => any` | `__dgram.socket.bind` | 📋 Planned | - |
| `socket.close([callback])` | `(...) => any` | `__dgram.socket.close` | 📋 Planned | - |
| `socket.connect(port[, address][, callback])` | `(...) => any` | `__dgram.socket.connect` | 📋 Planned | - |
| `socket.disconnect()` | `(...) => any` | `__dgram.socket.disconnect` | 📋 Planned | - |
| `socket.dropMembership(multicastAddress[, multicastInterface])` | `(...) => any` | `__dgram.socket.dropMembership` | 📋 Planned | - |
| `socket.dropSourceSpecificMembership(sourceAddress, groupAddress[, multicastInterface])` | `(...) => any` | `__dgram.socket.dropSourceSpecificMembership` | 📋 Planned | - |
| `socket.getRecvBufferSize()` | `(...) => any` | `__dgram.socket.getRecvBufferSize` | 📋 Planned | - |
| `socket.getSendBufferSize()` | `(...) => any` | `__dgram.socket.getSendBufferSize` | 📋 Planned | - |
| `socket.getSendQueueCount()` | `(...) => any` | `__dgram.socket.getSendQueueCount` | 📋 Planned | - |
| `socket.getSendQueueSize()` | `(...) => any` | `__dgram.socket.getSendQueueSize` | 📋 Planned | - |
| `socket.ref()` | `(...) => any` | `__dgram.socket.ref` | 📋 Planned | - |
| `socket.remoteAddress()` | `(...) => any` | `__dgram.socket.remoteAddress` | 📋 Planned | - |
| `socket.send(msg[, offset, length][, port][, address][, callback])` | `(...) => any` | `__dgram.socket.send` | 📋 Planned | - |
| `socket.setBroadcast(flag)` | `(...) => any` | `__dgram.socket.setBroadcast` | 📋 Planned | - |
| `socket.setMulticastInterface(multicastInterface)` | `(...) => any` | `__dgram.socket.setMulticastInterface` | 📋 Planned | - |
| `socket.setMulticastLoopback(flag)` | `(...) => any` | `__dgram.socket.setMulticastLoopback` | 📋 Planned | - |
| `socket.setMulticastTTL(ttl)` | `(...) => any` | `__dgram.socket.setMulticastTTL` | 📋 Planned | - |
| `socket.setRecvBufferSize(size)` | `(...) => any` | `__dgram.socket.setRecvBufferSize` | 📋 Planned | - |
| `socket.setSendBufferSize(size)` | `(...) => any` | `__dgram.socket.setSendBufferSize` | 📋 Planned | - |
| `socket.setTTL(ttl)` | `(...) => any` | `__dgram.socket.setTTL` | 📋 Planned | - |
| `socket.unref()` | `(...) => any` | `__dgram.socket.unref` | 📋 Planned | - |
| `socket[Symbol.asyncDispose]()` | `(...) => any` | `__dgram.socket[Symbol.asyncDispose]` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `dgram` are organized per API under `internal/compiler/testdata/corpus/dgram/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/dgram/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
