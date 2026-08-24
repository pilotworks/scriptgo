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
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `BlockList.isBlockList(value)` | `(...) => any` | `__net.BlockList.isBlockList` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `SocketAddress.parse(input)` | `(...) => any` | `__net.SocketAddress.parse` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `address` | `any` | `__net.address` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `autoSelectFamilyAttemptedAddresses` | `any` | `__net.autoSelectFamilyAttemptedAddresses` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.addAddress(address[, type])` | `(...) => any` | `__net.blockList.addAddress` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.addRange(start, end[, type])` | `(...) => any` | `__net.blockList.addRange` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.addSubnet(net, prefix[, type])` | `(...) => any` | `__net.blockList.addSubnet` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.check(address[, type])` | `(...) => any` | `__net.blockList.check` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.fromJSON(value)` | `(...) => any` | `__net.blockList.fromJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `blockList.toJSON()` | `(...) => any` | `__net.blockList.toJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `bufferSize` | `any` | `__net.bufferSize` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `bytesRead` | `any` | `__net.bytesRead` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `bytesWritten` | `any` | `__net.bytesWritten` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `connecting` | `any` | `__net.connecting` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `destroyed` | `any` | `__net.destroyed` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `dropMaxConnection` | `any` | `__net.dropMaxConnection` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `family` | `any` | `__net.family` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `flowlabel` | `any` | `__net.flowlabel` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `listening` | `any` | `__net.listening` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `localAddress` | `any` | `__net.localAddress` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `localFamily` | `any` | `__net.localFamily` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `localPort` | `any` | `__net.localPort` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `maxConnections` | `any` | `__net.maxConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.BlockList` | `(...) => any` | `__net.net.BlockList` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.Server` | `(...) => any` | `__net.net.Server` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.Socket` | `(...) => any` | `__net.net.Socket` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.SocketAddress` | `(...) => any` | `__net.net.SocketAddress` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.connect()` | `(...) => any` | `__net.net.connect` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.createConnection()` | `(...) => any` | `__net.net.createConnection` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.createServer([options][, connectionListener])` | `(...) => any` | `__net.net.createServer` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.getDefaultAutoSelectFamily()` | `(...) => any` | `__net.net.getDefaultAutoSelectFamily` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.getDefaultAutoSelectFamilyAttemptTimeout()` | `(...) => any` | `__net.net.getDefaultAutoSelectFamilyAttemptTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.isIP(input)` | `(...) => any` | `__net.net.isIP` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.isIPv4(input)` | `(...) => any` | `__net.net.isIPv4` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.isIPv6(input)` | `(...) => any` | `__net.net.isIPv6` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.setDefaultAutoSelectFamily(value)` | `(...) => any` | `__net.net.setDefaultAutoSelectFamily` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `net.setDefaultAutoSelectFamilyAttemptTimeout(value)` | `(...) => any` | `__net.net.setDefaultAutoSelectFamilyAttemptTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `pending` | `any` | `__net.pending` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `port` | `any` | `__net.port` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `readyState` | `any` | `__net.readyState` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `remoteAddress` | `any` | `__net.remoteAddress` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `remoteFamily` | `any` | `__net.remoteFamily` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `remotePort` | `any` | `__net.remotePort` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `rules` | `any` | `__net.rules` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.address()` | `(...) => any` | `__net.server.address` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.close([callback])` | `(...) => any` | `__net.server.close` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.getConnections(callback)` | `(...) => any` | `__net.server.getConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.listen()` | `(...) => any` | `__net.server.listen` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.ref()` | `(...) => any` | `__net.server.ref` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server.unref()` | `(...) => any` | `__net.server.unref` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__net.server[Symbol.asyncDispose]` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.address()` | `(...) => any` | `__net.socket.address` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.connect()` | `(...) => any` | `__net.socket.connect` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.destroy([error])` | `(...) => any` | `__net.socket.destroy` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.destroySoon()` | `(...) => any` | `__net.socket.destroySoon` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.end([data[, encoding]][, callback])` | `(...) => any` | `__net.socket.end` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.pause()` | `(...) => any` | `__net.socket.pause` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.ref()` | `(...) => any` | `__net.socket.ref` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.resetAndDestroy()` | `(...) => any` | `__net.socket.resetAndDestroy` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.resume()` | `(...) => any` | `__net.socket.resume` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.setEncoding([encoding])` | `(...) => any` | `__net.socket.setEncoding` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.setKeepAlive([enable][, initialDelay])` | `(...) => any` | `__net.socket.setKeepAlive` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.setNoDelay([noDelay])` | `(...) => any` | `__net.socket.setNoDelay` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.setTimeout(timeout[, callback])` | `(...) => any` | `__net.socket.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.unref()` | `(...) => any` | `__net.socket.unref` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `socket.write(data[, encoding][, callback])` | `(...) => any` | `__net.socket.write` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |
| `timeout` | `any` | `__net.timeout` | ✅ Done | `internal/compiler/testdata/corpus/api/net.ts` |

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
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/net/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
