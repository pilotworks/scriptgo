# HTTP/2 Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:http_2`  
> **Specification Reference**: [Node.js 22 LTS HTTP/2 Documentation](https://nodejs.org/docs/latest-v22.x/api/http_2.html)  
> **Type Definition Source**: [@types/node/http_2.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-http_2-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:http_2`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `ClientHttp2Session` | `(...) => any` | `__http_2.ClientHttp2Session` | 📋 Planned | - |
| `ClientHttp2Stream` | `(...) => any` | `__http_2.ClientHttp2Stream` | 📋 Planned | - |
| `Http2SecureServer` | `(...) => any` | `__http_2.Http2SecureServer` | 📋 Planned | - |
| `Http2Server` | `(...) => any` | `__http_2.Http2Server` | 📋 Planned | - |
| `Http2Session` | `(...) => any` | `__http_2.Http2Session` | 📋 Planned | - |
| `Http2Stream` | `(...) => any` | `__http_2.Http2Stream` | 📋 Planned | - |
| `ServerHttp2Session` | `(...) => any` | `__http_2.ServerHttp2Session` | 📋 Planned | - |
| `ServerHttp2Stream` | `(...) => any` | `__http_2.ServerHttp2Stream` | 📋 Planned | - |
| `aborted` | `any` | `__http_2.aborted` | 📋 Planned | - |
| `alpnProtocol` | `any` | `__http_2.alpnProtocol` | 📋 Planned | - |
| `authority` | `any` | `__http_2.authority` | 📋 Planned | - |
| `bufferSize` | `any` | `__http_2.bufferSize` | 📋 Planned | - |
| `clienthttp2session.request(headers[, options])` | `(...) => any` | `__http_2.clienthttp2session.request` | 📋 Planned | - |
| `closed` | `any` | `__http_2.closed` | 📋 Planned | - |
| `complete` | `any` | `__http_2.complete` | 📋 Planned | - |
| `connecting` | `any` | `__http_2.connecting` | 📋 Planned | - |
| `connection` | `any` | `__http_2.connection` | 📋 Planned | - |
| `destroyed` | `any` | `__http_2.destroyed` | 📋 Planned | - |
| `encrypted` | `any` | `__http_2.encrypted` | 📋 Planned | - |
| `endAfterHeaders` | `any` | `__http_2.endAfterHeaders` | 📋 Planned | - |
| `finished` | `any` | `__http_2.finished` | 📋 Planned | - |
| `headers` | `any` | `__http_2.headers` | 📋 Planned | - |
| `headersSent` | `any` | `__http_2.headersSent` | 📋 Planned | - |
| `http2.Http2ServerRequest` | `(...) => any` | `__http_2.http2.Http2ServerRequest` | 📋 Planned | - |
| `http2.Http2ServerResponse` | `(...) => any` | `__http_2.http2.Http2ServerResponse` | 📋 Planned | - |
| `http2.connect(authority[, options][, listener])` | `(...) => any` | `__http_2.http2.connect` | 📋 Planned | - |
| `http2.constants` | `any` | `__http_2.http2.constants` | 📋 Planned | - |
| `http2.createSecureServer(options[, onRequestHandler])` | `(...) => any` | `__http_2.http2.createSecureServer` | 📋 Planned | - |
| `http2.createServer([options][, onRequestHandler])` | `(...) => any` | `__http_2.http2.createServer` | 📋 Planned | - |
| `http2.getDefaultSettings()` | `(...) => any` | `__http_2.http2.getDefaultSettings` | 📋 Planned | - |
| `http2.getPackedSettings([settings])` | `(...) => any` | `__http_2.http2.getPackedSettings` | 📋 Planned | - |
| `http2.getUnpackedSettings(buf)` | `(...) => any` | `__http_2.http2.getUnpackedSettings` | 📋 Planned | - |
| `http2.performServerHandshake(socket[, options])` | `(...) => any` | `__http_2.http2.performServerHandshake` | 📋 Planned | - |
| `http2session.close([callback])` | `(...) => any` | `__http_2.http2session.close` | 📋 Planned | - |
| `http2session.destroy([error][, code])` | `(...) => any` | `__http_2.http2session.destroy` | 📋 Planned | - |
| `http2session.goaway([code[, lastStreamID[, opaqueData]]])` | `(...) => any` | `__http_2.http2session.goaway` | 📋 Planned | - |
| `http2session.ping([payload, ]callback)` | `(...) => any` | `__http_2.http2session.ping` | 📋 Planned | - |
| `http2session.ref()` | `(...) => any` | `__http_2.http2session.ref` | 📋 Planned | - |
| `http2session.setLocalWindowSize(windowSize)` | `(...) => any` | `__http_2.http2session.setLocalWindowSize` | 📋 Planned | - |
| `http2session.setTimeout(msecs, callback)` | `(...) => any` | `__http_2.http2session.setTimeout` | 📋 Planned | - |
| `http2session.settings([settings][, callback])` | `(...) => any` | `__http_2.http2session.settings` | 📋 Planned | - |
| `http2session.state` | `any` | `__http_2.http2session.state` | 📋 Planned | - |
| `http2session.unref()` | `(...) => any` | `__http_2.http2session.unref` | 📋 Planned | - |
| `http2stream.additionalHeaders(headers)` | `(...) => any` | `__http_2.http2stream.additionalHeaders` | 📋 Planned | - |
| `http2stream.close(code[, callback])` | `(...) => any` | `__http_2.http2stream.close` | 📋 Planned | - |
| `http2stream.priority(options)` | `(...) => any` | `__http_2.http2stream.priority` | 📋 Planned | - |
| `http2stream.pushStream(headers[, options], callback)` | `(...) => any` | `__http_2.http2stream.pushStream` | 📋 Planned | - |
| `http2stream.respond([headers[, options]])` | `(...) => any` | `__http_2.http2stream.respond` | 📋 Planned | - |
| `http2stream.respondWithFD(fd[, headers[, options]])` | `(...) => any` | `__http_2.http2stream.respondWithFD` | 📋 Planned | - |
| `http2stream.respondWithFile(path[, headers[, options]])` | `(...) => any` | `__http_2.http2stream.respondWithFile` | 📋 Planned | - |
| `http2stream.sendTrailers(headers)` | `(...) => any` | `__http_2.http2stream.sendTrailers` | 📋 Planned | - |
| `http2stream.setTimeout(msecs, callback)` | `(...) => any` | `__http_2.http2stream.setTimeout` | 📋 Planned | - |
| `http2stream.state` | `any` | `__http_2.http2stream.state` | 📋 Planned | - |
| `httpVersion` | `any` | `__http_2.httpVersion` | 📋 Planned | - |
| `id` | `any` | `__http_2.id` | 📋 Planned | - |
| `localSettings` | `any` | `__http_2.localSettings` | 📋 Planned | - |
| `method` | `any` | `__http_2.method` | 📋 Planned | - |
| `originSet` | `any` | `__http_2.originSet` | 📋 Planned | - |
| `pending` | `any` | `__http_2.pending` | 📋 Planned | - |
| `pendingSettingsAck` | `any` | `__http_2.pendingSettingsAck` | 📋 Planned | - |
| `pushAllowed` | `any` | `__http_2.pushAllowed` | 📋 Planned | - |
| `rawHeaders` | `any` | `__http_2.rawHeaders` | 📋 Planned | - |
| `rawTrailers` | `any` | `__http_2.rawTrailers` | 📋 Planned | - |
| `remoteSettings` | `any` | `__http_2.remoteSettings` | 📋 Planned | - |
| `req` | `any` | `__http_2.req` | 📋 Planned | - |
| `request.destroy([error])` | `(...) => any` | `__http_2.request.destroy` | 📋 Planned | - |
| `request.setTimeout(msecs, callback)` | `(...) => any` | `__http_2.request.setTimeout` | 📋 Planned | - |
| `response.addTrailers(headers)` | `(...) => any` | `__http_2.response.addTrailers` | 📋 Planned | - |
| `response.appendHeader(name, value)` | `(...) => any` | `__http_2.response.appendHeader` | 📋 Planned | - |
| `response.createPushResponse(headers, callback)` | `(...) => any` | `__http_2.response.createPushResponse` | 📋 Planned | - |
| `response.end([data[, encoding]][, callback])` | `(...) => any` | `__http_2.response.end` | 📋 Planned | - |
| `response.getHeader(name)` | `(...) => any` | `__http_2.response.getHeader` | 📋 Planned | - |
| `response.getHeaderNames()` | `(...) => any` | `__http_2.response.getHeaderNames` | 📋 Planned | - |
| `response.getHeaders()` | `(...) => any` | `__http_2.response.getHeaders` | 📋 Planned | - |
| `response.hasHeader(name)` | `(...) => any` | `__http_2.response.hasHeader` | 📋 Planned | - |
| `response.removeHeader(name)` | `(...) => any` | `__http_2.response.removeHeader` | 📋 Planned | - |
| `response.setHeader(name, value)` | `(...) => any` | `__http_2.response.setHeader` | 📋 Planned | - |
| `response.setTimeout(msecs[, callback])` | `(...) => any` | `__http_2.response.setTimeout` | 📋 Planned | - |
| `response.write(chunk[, encoding][, callback])` | `(...) => any` | `__http_2.response.write` | 📋 Planned | - |
| `response.writeContinue()` | `(...) => any` | `__http_2.response.writeContinue` | 📋 Planned | - |
| `response.writeEarlyHints(hints)` | `(...) => any` | `__http_2.response.writeEarlyHints` | 📋 Planned | - |
| `response.writeHead(statusCode[, statusMessage][, headers])` | `(...) => any` | `__http_2.response.writeHead` | 📋 Planned | - |
| `rstCode` | `any` | `__http_2.rstCode` | 📋 Planned | - |
| `scheme` | `any` | `__http_2.scheme` | 📋 Planned | - |
| `sendDate` | `any` | `__http_2.sendDate` | 📋 Planned | - |
| `sensitiveHeaders` | `any` | `__http_2.sensitiveHeaders` | 📋 Planned | - |
| `sentHeaders` | `any` | `__http_2.sentHeaders` | 📋 Planned | - |
| `sentInfoHeaders` | `any` | `__http_2.sentInfoHeaders` | 📋 Planned | - |
| `sentTrailers` | `any` | `__http_2.sentTrailers` | 📋 Planned | - |
| `server.close([callback])` | `(...) => any` | `__http_2.server.close` | 📋 Planned | - |
| `server.setTimeout([msecs][, callback])` | `(...) => any` | `__http_2.server.setTimeout` | 📋 Planned | - |
| `server.updateSettings([settings])` | `(...) => any` | `__http_2.server.updateSettings` | 📋 Planned | - |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__http_2.server[Symbol.asyncDispose]` | 📋 Planned | - |
| `serverhttp2session.altsvc(alt, originOrStream)` | `(...) => any` | `__http_2.serverhttp2session.altsvc` | 📋 Planned | - |
| `serverhttp2session.origin(...origins)` | `(...) => any` | `__http_2.serverhttp2session.origin` | 📋 Planned | - |
| `session` | `any` | `__http_2.session` | 📋 Planned | - |
| `socket` | `any` | `__http_2.socket` | 📋 Planned | - |
| `statusCode` | `any` | `__http_2.statusCode` | 📋 Planned | - |
| `statusMessage` | `any` | `__http_2.statusMessage` | 📋 Planned | - |
| `stream` | `any` | `__http_2.stream` | 📋 Planned | - |
| `timeout` | `any` | `__http_2.timeout` | 📋 Planned | - |
| `trailers` | `any` | `__http_2.trailers` | 📋 Planned | - |
| `type` | `any` | `__http_2.type` | 📋 Planned | - |
| `url` | `any` | `__http_2.url` | 📋 Planned | - |
| `writableEnded` | `any` | `__http_2.writableEnded` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `http_2` are organized per API under `internal/compiler/testdata/corpus/http_2/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/http_2/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
