# HTTP Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:http`  
> **Specification Reference**: [Node.js 22 LTS HTTP Documentation](https://nodejs.org/docs/latest-v22.x/api/http.html)  
> **Type Definition Source**: [@types/node/http.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-http-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:http`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `METHODS` | `any` | `__http.METHODS` | ✅ Done | `internal/compiler/testdata/corpus/api/http/METHODS/` |
| `STATUS_CODES` | `any` | `__http.STATUS_CODES` | 📋 Planned | - |
| `WebSocket` | `(...) => any` | `__http.WebSocket` | 📋 Planned | - |
| `aborted` | `any` | `__http.aborted` | 📋 Planned | - |
| `agent.createConnection(options[, callback])` | `(...) => any` | `__http.agent.createConnection` | 📋 Planned | - |
| `agent.destroy()` | `(...) => any` | `__http.agent.destroy` | 📋 Planned | - |
| `agent.getName([options])` | `(...) => any` | `__http.agent.getName` | 📋 Planned | - |
| `agent.keepSocketAlive(socket)` | `(...) => any` | `__http.agent.keepSocketAlive` | 📋 Planned | - |
| `agent.reuseSocket(socket, request)` | `(...) => any` | `__http.agent.reuseSocket` | 📋 Planned | - |
| `complete` | `any` | `__http.complete` | 📋 Planned | - |
| `connection` | `any` | `__http.connection` | 📋 Planned | - |
| `finished` | `any` | `__http.finished` | 📋 Planned | - |
| `freeSockets` | `any` | `__http.freeSockets` | 📋 Planned | - |
| `globalAgent` | `any` | `__http.globalAgent` | 📋 Planned | - |
| `headers` | `any` | `__http.headers` | 📋 Planned | - |
| `headersDistinct` | `any` | `__http.headersDistinct` | 📋 Planned | - |
| `headersSent` | `any` | `__http.headersSent` | 📋 Planned | - |
| `headersTimeout` | `any` | `__http.headersTimeout` | 📋 Planned | - |
| `host` | `any` | `__http.host` | 📋 Planned | - |
| `http.Agent` | `(...) => any` | `__http.http.Agent` | 📋 Planned | - |
| `http.ClientRequest` | `(...) => any` | `__http.http.ClientRequest` | 📋 Planned | - |
| `http.IncomingMessage` | `(...) => any` | `__http.http.IncomingMessage` | 📋 Planned | - |
| `http.OutgoingMessage` | `(...) => any` | `__http.http.OutgoingMessage` | 📋 Planned | - |
| `http.Server` | `(...) => any` | `__http.http.Server` | 📋 Planned | - |
| `http.ServerResponse` | `(...) => any` | `__http.http.ServerResponse` | 📋 Planned | - |
| `http.createServer([options][, requestListener])` | `(...) => any` | `__http.http.createServer` | 📋 Planned | - |
| `http.get(options[, callback])` | `(...) => any` | `__http.http.get` | 📋 Planned | - |
| `http.get(url[, options][, callback])` | `(...) => any` | `__http.http.get` | 📋 Planned | - |
| `http.request(options[, callback])` | `(...) => any` | `__http.http.request` | 📋 Planned | - |
| `http.request(url[, options][, callback])` | `(...) => any` | `__http.http.request` | 📋 Planned | - |
| `http.setMaxIdleHTTPParsers(max)` | `(...) => any` | `__http.http.setMaxIdleHTTPParsers` | 📋 Planned | - |
| `http.validateHeaderName(name[, label])` | `(...) => any` | `__http.http.validateHeaderName` | 📋 Planned | - |
| `http.validateHeaderValue(name, value)` | `(...) => any` | `__http.http.validateHeaderValue` | 📋 Planned | - |
| `httpVersion` | `any` | `__http.httpVersion` | 📋 Planned | - |
| `keepAliveTimeout` | `any` | `__http.keepAliveTimeout` | 📋 Planned | - |
| `keepAliveTimeoutBuffer` | `any` | `__http.keepAliveTimeoutBuffer` | 📋 Planned | - |
| `listening` | `any` | `__http.listening` | 📋 Planned | - |
| `maxFreeSockets` | `any` | `__http.maxFreeSockets` | 📋 Planned | - |
| `maxHeaderSize` | `any` | `__http.maxHeaderSize` | 📋 Planned | - |
| `maxHeadersCount` | `any` | `__http.maxHeadersCount` | 📋 Planned | - |
| `maxRequestsPerSocket` | `any` | `__http.maxRequestsPerSocket` | 📋 Planned | - |
| `maxSockets` | `any` | `__http.maxSockets` | 📋 Planned | - |
| `maxTotalSockets` | `any` | `__http.maxTotalSockets` | 📋 Planned | - |
| `message.connection` | `any` | `__http.message.connection` | 📋 Planned | - |
| `message.destroy([error])` | `(...) => any` | `__http.message.destroy` | 📋 Planned | - |
| `message.setTimeout(msecs[, callback])` | `(...) => any` | `__http.message.setTimeout` | 📋 Planned | - |
| `method` | `any` | `__http.method` | 📋 Planned | - |
| `outgoingMessage.addTrailers(headers)` | `(...) => any` | `__http.outgoingMessage.addTrailers` | 📋 Planned | - |
| `outgoingMessage.appendHeader(name, value)` | `(...) => any` | `__http.outgoingMessage.appendHeader` | 📋 Planned | - |
| `outgoingMessage.connection` | `any` | `__http.outgoingMessage.connection` | 📋 Planned | - |
| `outgoingMessage.cork()` | `(...) => any` | `__http.outgoingMessage.cork` | 📋 Planned | - |
| `outgoingMessage.destroy([error])` | `(...) => any` | `__http.outgoingMessage.destroy` | 📋 Planned | - |
| `outgoingMessage.end(chunk[, encoding][, callback])` | `(...) => any` | `__http.outgoingMessage.end` | 📋 Planned | - |
| `outgoingMessage.flushHeaders()` | `(...) => any` | `__http.outgoingMessage.flushHeaders` | 📋 Planned | - |
| `outgoingMessage.getHeader(name)` | `(...) => any` | `__http.outgoingMessage.getHeader` | 📋 Planned | - |
| `outgoingMessage.getHeaderNames()` | `(...) => any` | `__http.outgoingMessage.getHeaderNames` | 📋 Planned | - |
| `outgoingMessage.getHeaders()` | `(...) => any` | `__http.outgoingMessage.getHeaders` | 📋 Planned | - |
| `outgoingMessage.hasHeader(name)` | `(...) => any` | `__http.outgoingMessage.hasHeader` | 📋 Planned | - |
| `outgoingMessage.pipe()` | `(...) => any` | `__http.outgoingMessage.pipe` | 📋 Planned | - |
| `outgoingMessage.removeHeader(name)` | `(...) => any` | `__http.outgoingMessage.removeHeader` | 📋 Planned | - |
| `outgoingMessage.setHeader(name, value)` | `(...) => any` | `__http.outgoingMessage.setHeader` | 📋 Planned | - |
| `outgoingMessage.setHeaders(headers)` | `(...) => any` | `__http.outgoingMessage.setHeaders` | 📋 Planned | - |
| `outgoingMessage.setTimeout(msecs[, callback])` | `(...) => any` | `__http.outgoingMessage.setTimeout` | 📋 Planned | - |
| `outgoingMessage.uncork()` | `(...) => any` | `__http.outgoingMessage.uncork` | 📋 Planned | - |
| `outgoingMessage.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.outgoingMessage.write` | 📋 Planned | - |
| `path` | `any` | `__http.path` | 📋 Planned | - |
| `protocol` | `any` | `__http.protocol` | 📋 Planned | - |
| `rawHeaders` | `any` | `__http.rawHeaders` | 📋 Planned | - |
| `rawTrailers` | `any` | `__http.rawTrailers` | 📋 Planned | - |
| `req` | `any` | `__http.req` | 📋 Planned | - |
| `request.abort()` | `(...) => any` | `__http.request.abort` | 📋 Planned | - |
| `request.cork()` | `(...) => any` | `__http.request.cork` | 📋 Planned | - |
| `request.destroy([error])` | `(...) => any` | `__http.request.destroy` | 📋 Planned | - |
| `request.end([data[, encoding]][, callback])` | `(...) => any` | `__http.request.end` | 📋 Planned | - |
| `request.flushHeaders()` | `(...) => any` | `__http.request.flushHeaders` | 📋 Planned | - |
| `request.getHeader(name)` | `(...) => any` | `__http.request.getHeader` | 📋 Planned | - |
| `request.getHeaderNames()` | `(...) => any` | `__http.request.getHeaderNames` | 📋 Planned | - |
| `request.getHeaders()` | `(...) => any` | `__http.request.getHeaders` | 📋 Planned | - |
| `request.getRawHeaderNames()` | `(...) => any` | `__http.request.getRawHeaderNames` | 📋 Planned | - |
| `request.hasHeader(name)` | `(...) => any` | `__http.request.hasHeader` | 📋 Planned | - |
| `request.removeHeader(name)` | `(...) => any` | `__http.request.removeHeader` | 📋 Planned | - |
| `request.setHeader(name, value)` | `(...) => any` | `__http.request.setHeader` | 📋 Planned | - |
| `request.setNoDelay([noDelay])` | `(...) => any` | `__http.request.setNoDelay` | 📋 Planned | - |
| `request.setSocketKeepAlive([enable][, initialDelay])` | `(...) => any` | `__http.request.setSocketKeepAlive` | 📋 Planned | - |
| `request.setTimeout(timeout[, callback])` | `(...) => any` | `__http.request.setTimeout` | 📋 Planned | - |
| `request.uncork()` | `(...) => any` | `__http.request.uncork` | 📋 Planned | - |
| `request.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.request.write` | 📋 Planned | - |
| `requestTimeout` | `any` | `__http.requestTimeout` | 📋 Planned | - |
| `requests` | `any` | `__http.requests` | 📋 Planned | - |
| `response.addTrailers(headers)` | `(...) => any` | `__http.response.addTrailers` | 📋 Planned | - |
| `response.cork()` | `(...) => any` | `__http.response.cork` | 📋 Planned | - |
| `response.end([data[, encoding]][, callback])` | `(...) => any` | `__http.response.end` | 📋 Planned | - |
| `response.flushHeaders()` | `(...) => any` | `__http.response.flushHeaders` | 📋 Planned | - |
| `response.getHeader(name)` | `(...) => any` | `__http.response.getHeader` | 📋 Planned | - |
| `response.getHeaderNames()` | `(...) => any` | `__http.response.getHeaderNames` | 📋 Planned | - |
| `response.getHeaders()` | `(...) => any` | `__http.response.getHeaders` | 📋 Planned | - |
| `response.hasHeader(name)` | `(...) => any` | `__http.response.hasHeader` | 📋 Planned | - |
| `response.removeHeader(name)` | `(...) => any` | `__http.response.removeHeader` | 📋 Planned | - |
| `response.setHeader(name, value)` | `(...) => any` | `__http.response.setHeader` | 📋 Planned | - |
| `response.setTimeout(msecs[, callback])` | `(...) => any` | `__http.response.setTimeout` | 📋 Planned | - |
| `response.uncork()` | `(...) => any` | `__http.response.uncork` | 📋 Planned | - |
| `response.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.response.write` | 📋 Planned | - |
| `response.writeContinue()` | `(...) => any` | `__http.response.writeContinue` | 📋 Planned | - |
| `response.writeEarlyHints(hints[, callback])` | `(...) => any` | `__http.response.writeEarlyHints` | 📋 Planned | - |
| `response.writeHead(statusCode[, statusMessage][, headers])` | `(...) => any` | `__http.response.writeHead` | 📋 Planned | - |
| `response.writeProcessing()` | `(...) => any` | `__http.response.writeProcessing` | 📋 Planned | - |
| `reusedSocket` | `any` | `__http.reusedSocket` | 📋 Planned | - |
| `sendDate` | `any` | `__http.sendDate` | 📋 Planned | - |
| `server.close([callback])` | `(...) => any` | `__http.server.close` | 📋 Planned | - |
| `server.closeAllConnections()` | `(...) => any` | `__http.server.closeAllConnections` | 📋 Planned | - |
| `server.closeIdleConnections()` | `(...) => any` | `__http.server.closeIdleConnections` | 📋 Planned | - |
| `server.listen()` | `(...) => any` | `__http.server.listen` | 📋 Planned | - |
| `server.setTimeout([msecs][, callback])` | `(...) => any` | `__http.server.setTimeout` | 📋 Planned | - |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__http.server[Symbol.asyncDispose]` | 📋 Planned | - |
| `socket` | `any` | `__http.socket` | 📋 Planned | - |
| `sockets` | `any` | `__http.sockets` | 📋 Planned | - |
| `statusCode` | `any` | `__http.statusCode` | 📋 Planned | - |
| `statusMessage` | `any` | `__http.statusMessage` | 📋 Planned | - |
| `strictContentLength` | `any` | `__http.strictContentLength` | 📋 Planned | - |
| `timeout` | `any` | `__http.timeout` | 📋 Planned | - |
| `trailers` | `any` | `__http.trailers` | 📋 Planned | - |
| `trailersDistinct` | `any` | `__http.trailersDistinct` | 📋 Planned | - |
| `url` | `any` | `__http.url` | 📋 Planned | - |
| `writableCorked` | `any` | `__http.writableCorked` | 📋 Planned | - |
| `writableEnded` | `any` | `__http.writableEnded` | 📋 Planned | - |
| `writableFinished` | `any` | `__http.writableFinished` | 📋 Planned | - |
| `writableHighWaterMark` | `any` | `__http.writableHighWaterMark` | 📋 Planned | - |
| `writableLength` | `any` | `__http.writableLength` | 📋 Planned | - |
| `writableObjectMode` | `any` | `__http.writableObjectMode` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `http` are organized per API under `internal/compiler/testdata/corpus/http/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/http/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
