# Net Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:net`  
> **Specification Reference**: [Node.js 22 LTS Net Documentation](https://nodejs.org/docs/latest-v22.x/api/net.html)  
> **Type Definition Source**: [@types/node/net.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-net-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:net`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `BlockList.isBlockList(value)` | `(...) => any` | `__net.BlockList.isBlockList` | 📋 Planned | - |
| `SocketAddress.parse(input)` | `(...) => any` | `__net.SocketAddress.parse` | 📋 Planned | - |
| `address` | `any` | `__net.address` | 📋 Planned | - |
| `autoSelectFamilyAttemptedAddresses` | `any` | `__net.autoSelectFamilyAttemptedAddresses` | 📋 Planned | - |
| `blockList.addAddress(address[, type])` | `(...) => any` | `__net.blockList.addAddress` | 📋 Planned | - |
| `blockList.addRange(start, end[, type])` | `(...) => any` | `__net.blockList.addRange` | 📋 Planned | - |
| `blockList.addSubnet(net, prefix[, type])` | `(...) => any` | `__net.blockList.addSubnet` | 📋 Planned | - |
| `blockList.check(address[, type])` | `(...) => any` | `__net.blockList.check` | 📋 Planned | - |
| `blockList.fromJSON(value)` | `(...) => any` | `__net.blockList.fromJSON` | 📋 Planned | - |
| `blockList.toJSON()` | `(...) => any` | `__net.blockList.toJSON` | 📋 Planned | - |
| `bufferSize` | `any` | `__net.bufferSize` | 📋 Planned | - |
| `bytesRead` | `any` | `__net.bytesRead` | 📋 Planned | - |
| `bytesWritten` | `any` | `__net.bytesWritten` | 📋 Planned | - |
| `connecting` | `any` | `__net.connecting` | 📋 Planned | - |
| `destroyed` | `any` | `__net.destroyed` | 📋 Planned | - |
| `dropMaxConnection` | `any` | `__net.dropMaxConnection` | 📋 Planned | - |
| `family` | `any` | `__net.family` | 📋 Planned | - |
| `flowlabel` | `any` | `__net.flowlabel` | 📋 Planned | - |
| `listening` | `any` | `__net.listening` | 📋 Planned | - |
| `localAddress` | `any` | `__net.localAddress` | 📋 Planned | - |
| `localFamily` | `any` | `__net.localFamily` | 📋 Planned | - |
| `localPort` | `any` | `__net.localPort` | 📋 Planned | - |
| `maxConnections` | `any` | `__net.maxConnections` | 📋 Planned | - |
| `net.BlockList` | `(...) => any` | `__net.net.BlockList` | 📋 Planned | - |
| `net.Server` | `(...) => any` | `__net.net.Server` | 📋 Planned | - |
| `net.Socket` | `(...) => any` | `__net.net.Socket` | 📋 Planned | - |
| `net.SocketAddress` | `(...) => any` | `__net.net.SocketAddress` | 📋 Planned | - |
| `net.connect()` | `(...) => any` | `__net.net.connect` | 📋 Planned | - |
| `net.createConnection()` | `(...) => any` | `__net.net.createConnection` | 📋 Planned | - |
| `net.createServer([options][, connectionListener])` | `(...) => any` | `__net.net.createServer` | 📋 Planned | - |
| `net.getDefaultAutoSelectFamily()` | `(...) => any` | `__net.net.getDefaultAutoSelectFamily` | 📋 Planned | - |
| `net.getDefaultAutoSelectFamilyAttemptTimeout()` | `(...) => any` | `__net.net.getDefaultAutoSelectFamilyAttemptTimeout` | 📋 Planned | - |
| `net.isIP(input)` | `(...) => any` | `__net.net.isIP` | 📋 Planned | - |
| `net.isIPv4(input)` | `(...) => any` | `__net.net.isIPv4` | 📋 Planned | - |
| `net.isIPv6(input)` | `(...) => any` | `__net.net.isIPv6` | 📋 Planned | - |
| `net.setDefaultAutoSelectFamily(value)` | `(...) => any` | `__net.net.setDefaultAutoSelectFamily` | 📋 Planned | - |
| `net.setDefaultAutoSelectFamilyAttemptTimeout(value)` | `(...) => any` | `__net.net.setDefaultAutoSelectFamilyAttemptTimeout` | 📋 Planned | - |
| `pending` | `any` | `__net.pending` | 📋 Planned | - |
| `port` | `any` | `__net.port` | 📋 Planned | - |
| `readyState` | `any` | `__net.readyState` | 📋 Planned | - |
| `remoteAddress` | `any` | `__net.remoteAddress` | 📋 Planned | - |
| `remoteFamily` | `any` | `__net.remoteFamily` | 📋 Planned | - |
| `remotePort` | `any` | `__net.remotePort` | 📋 Planned | - |
| `rules` | `any` | `__net.rules` | 📋 Planned | - |
| `server.address()` | `(...) => any` | `__net.server.address` | 📋 Planned | - |
| `server.close([callback])` | `(...) => any` | `__net.server.close` | 📋 Planned | - |
| `server.getConnections(callback)` | `(...) => any` | `__net.server.getConnections` | 📋 Planned | - |
| `server.listen()` | `(...) => any` | `__net.server.listen` | 📋 Planned | - |
| `server.ref()` | `(...) => any` | `__net.server.ref` | 📋 Planned | - |
| `server.unref()` | `(...) => any` | `__net.server.unref` | 📋 Planned | - |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__net.server[Symbol.asyncDispose]` | 📋 Planned | - |
| `socket.address()` | `(...) => any` | `__net.socket.address` | 📋 Planned | - |
| `socket.connect()` | `(...) => any` | `__net.socket.connect` | 📋 Planned | - |
| `socket.destroy([error])` | `(...) => any` | `__net.socket.destroy` | 📋 Planned | - |
| `socket.destroySoon()` | `(...) => any` | `__net.socket.destroySoon` | 📋 Planned | - |
| `socket.end([data[, encoding]][, callback])` | `(...) => any` | `__net.socket.end` | 📋 Planned | - |
| `socket.pause()` | `(...) => any` | `__net.socket.pause` | 📋 Planned | - |
| `socket.ref()` | `(...) => any` | `__net.socket.ref` | 📋 Planned | - |
| `socket.resetAndDestroy()` | `(...) => any` | `__net.socket.resetAndDestroy` | 📋 Planned | - |
| `socket.resume()` | `(...) => any` | `__net.socket.resume` | 📋 Planned | - |
| `socket.setEncoding([encoding])` | `(...) => any` | `__net.socket.setEncoding` | 📋 Planned | - |
| `socket.setKeepAlive([enable][, initialDelay])` | `(...) => any` | `__net.socket.setKeepAlive` | 📋 Planned | - |
| `socket.setNoDelay([noDelay])` | `(...) => any` | `__net.socket.setNoDelay` | 📋 Planned | - |
| `socket.setTimeout(timeout[, callback])` | `(...) => any` | `__net.socket.setTimeout` | 📋 Planned | - |
| `socket.unref()` | `(...) => any` | `__net.socket.unref` | 📋 Planned | - |
| `socket.write(data[, encoding][, callback])` | `(...) => any` | `__net.socket.write` | 📋 Planned | - |
| `timeout` | `any` | `__net.timeout` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `net` are organized per API under `internal/compiler/testdata/corpus/net/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/net/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
