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
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `METHODS` | `any` | `__http.METHODS` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `STATUS_CODES` | `any` | `__http.STATUS_CODES` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `WebSocket` | `(...) => any` | `__http.WebSocket` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `aborted` | `any` | `__http.aborted` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `agent.createConnection(options[, callback])` | `(...) => any` | `__http.agent.createConnection` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `agent.destroy()` | `(...) => any` | `__http.agent.destroy` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `agent.getName([options])` | `(...) => any` | `__http.agent.getName` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `agent.keepSocketAlive(socket)` | `(...) => any` | `__http.agent.keepSocketAlive` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `agent.reuseSocket(socket, request)` | `(...) => any` | `__http.agent.reuseSocket` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `complete` | `any` | `__http.complete` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `connection` | `any` | `__http.connection` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `finished` | `any` | `__http.finished` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `freeSockets` | `any` | `__http.freeSockets` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `globalAgent` | `any` | `__http.globalAgent` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `headers` | `any` | `__http.headers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `headersDistinct` | `any` | `__http.headersDistinct` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `headersSent` | `any` | `__http.headersSent` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `headersTimeout` | `any` | `__http.headersTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `host` | `any` | `__http.host` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.Agent` | `(...) => any` | `__http.http.Agent` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.ClientRequest` | `(...) => any` | `__http.http.ClientRequest` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.IncomingMessage` | `(...) => any` | `__http.http.IncomingMessage` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.OutgoingMessage` | `(...) => any` | `__http.http.OutgoingMessage` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.Server` | `(...) => any` | `__http.http.Server` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.ServerResponse` | `(...) => any` | `__http.http.ServerResponse` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.createServer([options][, requestListener])` | `(...) => any` | `__http.http.createServer` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.get(options[, callback])` | `(...) => any` | `__http.http.get` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.get(url[, options][, callback])` | `(...) => any` | `__http.http.get` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.request(options[, callback])` | `(...) => any` | `__http.http.request` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.request(url[, options][, callback])` | `(...) => any` | `__http.http.request` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.setMaxIdleHTTPParsers(max)` | `(...) => any` | `__http.http.setMaxIdleHTTPParsers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.validateHeaderName(name[, label])` | `(...) => any` | `__http.http.validateHeaderName` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `http.validateHeaderValue(name, value)` | `(...) => any` | `__http.http.validateHeaderValue` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `httpVersion` | `any` | `__http.httpVersion` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `keepAliveTimeout` | `any` | `__http.keepAliveTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `keepAliveTimeoutBuffer` | `any` | `__http.keepAliveTimeoutBuffer` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `listening` | `any` | `__http.listening` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxFreeSockets` | `any` | `__http.maxFreeSockets` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxHeaderSize` | `any` | `__http.maxHeaderSize` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxHeadersCount` | `any` | `__http.maxHeadersCount` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxRequestsPerSocket` | `any` | `__http.maxRequestsPerSocket` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxSockets` | `any` | `__http.maxSockets` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `maxTotalSockets` | `any` | `__http.maxTotalSockets` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `message.connection` | `any` | `__http.message.connection` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `message.destroy([error])` | `(...) => any` | `__http.message.destroy` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `message.setTimeout(msecs[, callback])` | `(...) => any` | `__http.message.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `method` | `any` | `__http.method` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.addTrailers(headers)` | `(...) => any` | `__http.outgoingMessage.addTrailers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.appendHeader(name, value)` | `(...) => any` | `__http.outgoingMessage.appendHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.connection` | `any` | `__http.outgoingMessage.connection` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.cork()` | `(...) => any` | `__http.outgoingMessage.cork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.destroy([error])` | `(...) => any` | `__http.outgoingMessage.destroy` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.end(chunk[, encoding][, callback])` | `(...) => any` | `__http.outgoingMessage.end` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.flushHeaders()` | `(...) => any` | `__http.outgoingMessage.flushHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.getHeader(name)` | `(...) => any` | `__http.outgoingMessage.getHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.getHeaderNames()` | `(...) => any` | `__http.outgoingMessage.getHeaderNames` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.getHeaders()` | `(...) => any` | `__http.outgoingMessage.getHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.hasHeader(name)` | `(...) => any` | `__http.outgoingMessage.hasHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.pipe()` | `(...) => any` | `__http.outgoingMessage.pipe` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.removeHeader(name)` | `(...) => any` | `__http.outgoingMessage.removeHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.setHeader(name, value)` | `(...) => any` | `__http.outgoingMessage.setHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.setHeaders(headers)` | `(...) => any` | `__http.outgoingMessage.setHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.setTimeout(msecs[, callback])` | `(...) => any` | `__http.outgoingMessage.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.uncork()` | `(...) => any` | `__http.outgoingMessage.uncork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `outgoingMessage.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.outgoingMessage.write` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `path` | `any` | `__http.path` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `protocol` | `any` | `__http.protocol` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `rawHeaders` | `any` | `__http.rawHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `rawTrailers` | `any` | `__http.rawTrailers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `req` | `any` | `__http.req` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.abort()` | `(...) => any` | `__http.request.abort` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.cork()` | `(...) => any` | `__http.request.cork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.destroy([error])` | `(...) => any` | `__http.request.destroy` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.end([data[, encoding]][, callback])` | `(...) => any` | `__http.request.end` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.flushHeaders()` | `(...) => any` | `__http.request.flushHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.getHeader(name)` | `(...) => any` | `__http.request.getHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.getHeaderNames()` | `(...) => any` | `__http.request.getHeaderNames` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.getHeaders()` | `(...) => any` | `__http.request.getHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.getRawHeaderNames()` | `(...) => any` | `__http.request.getRawHeaderNames` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.hasHeader(name)` | `(...) => any` | `__http.request.hasHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.removeHeader(name)` | `(...) => any` | `__http.request.removeHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.setHeader(name, value)` | `(...) => any` | `__http.request.setHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.setNoDelay([noDelay])` | `(...) => any` | `__http.request.setNoDelay` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.setSocketKeepAlive([enable][, initialDelay])` | `(...) => any` | `__http.request.setSocketKeepAlive` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.setTimeout(timeout[, callback])` | `(...) => any` | `__http.request.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.uncork()` | `(...) => any` | `__http.request.uncork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `request.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.request.write` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `requestTimeout` | `any` | `__http.requestTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `requests` | `any` | `__http.requests` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.addTrailers(headers)` | `(...) => any` | `__http.response.addTrailers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.cork()` | `(...) => any` | `__http.response.cork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.end([data[, encoding]][, callback])` | `(...) => any` | `__http.response.end` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.flushHeaders()` | `(...) => any` | `__http.response.flushHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.getHeader(name)` | `(...) => any` | `__http.response.getHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.getHeaderNames()` | `(...) => any` | `__http.response.getHeaderNames` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.getHeaders()` | `(...) => any` | `__http.response.getHeaders` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.hasHeader(name)` | `(...) => any` | `__http.response.hasHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.removeHeader(name)` | `(...) => any` | `__http.response.removeHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.setHeader(name, value)` | `(...) => any` | `__http.response.setHeader` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.setTimeout(msecs[, callback])` | `(...) => any` | `__http.response.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.uncork()` | `(...) => any` | `__http.response.uncork` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.write(chunk[, encoding][, callback])` | `(...) => any` | `__http.response.write` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.writeContinue()` | `(...) => any` | `__http.response.writeContinue` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.writeEarlyHints(hints[, callback])` | `(...) => any` | `__http.response.writeEarlyHints` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.writeHead(statusCode[, statusMessage][, headers])` | `(...) => any` | `__http.response.writeHead` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `response.writeProcessing()` | `(...) => any` | `__http.response.writeProcessing` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `reusedSocket` | `any` | `__http.reusedSocket` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `sendDate` | `any` | `__http.sendDate` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server.close([callback])` | `(...) => any` | `__http.server.close` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server.closeAllConnections()` | `(...) => any` | `__http.server.closeAllConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server.closeIdleConnections()` | `(...) => any` | `__http.server.closeIdleConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server.listen()` | `(...) => any` | `__http.server.listen` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server.setTimeout([msecs][, callback])` | `(...) => any` | `__http.server.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__http.server[Symbol.asyncDispose]` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `socket` | `any` | `__http.socket` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `sockets` | `any` | `__http.sockets` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `statusCode` | `any` | `__http.statusCode` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `statusMessage` | `any` | `__http.statusMessage` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `strictContentLength` | `any` | `__http.strictContentLength` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `timeout` | `any` | `__http.timeout` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `trailers` | `any` | `__http.trailers` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `trailersDistinct` | `any` | `__http.trailersDistinct` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `url` | `any` | `__http.url` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableCorked` | `any` | `__http.writableCorked` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableEnded` | `any` | `__http.writableEnded` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableFinished` | `any` | `__http.writableFinished` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableHighWaterMark` | `any` | `__http.writableHighWaterMark` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableLength` | `any` | `__http.writableLength` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |
| `writableObjectMode` | `any` | `__http.writableObjectMode` | ✅ Done | `internal/compiler/testdata/corpus/api/http.ts` |

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
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/http/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
