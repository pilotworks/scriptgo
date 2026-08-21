# TLS (SSL) Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:tls_(ssl)`  
> **Specification Reference**: [Node.js 22 LTS TLS (SSL) Documentation](https://nodejs.org/docs/latest-v22.x/api/tls_(ssl).html)  
> **Type Definition Source**: [@types/node/tls_(ssl).d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-tls_(ssl)-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:tls_(ssl)`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `DEFAULT_CIPHERS` | `any` | `__tls_(ssl).DEFAULT_CIPHERS` | 📋 Planned | - |
| `DEFAULT_MAX_VERSION` | `any` | `__tls_(ssl).DEFAULT_MAX_VERSION` | 📋 Planned | - |
| `DEFAULT_MIN_VERSION` | `any` | `__tls_(ssl).DEFAULT_MIN_VERSION` | 📋 Planned | - |
| `authorized` | `any` | `__tls_(ssl).authorized` | 📋 Planned | - |
| `localAddress` | `any` | `__tls_(ssl).localAddress` | 📋 Planned | - |
| `localPort` | `any` | `__tls_(ssl).localPort` | 📋 Planned | - |
| `remoteAddress` | `any` | `__tls_(ssl).remoteAddress` | 📋 Planned | - |
| `remoteFamily` | `any` | `__tls_(ssl).remoteFamily` | 📋 Planned | - |
| `remotePort` | `any` | `__tls_(ssl).remotePort` | 📋 Planned | - |
| `rootCertificates` | `any` | `__tls_(ssl).rootCertificates` | 📋 Planned | - |
| `server.addContext(hostname, context)` | `(...) => any` | `__tls_(ssl).server.addContext` | 📋 Planned | - |
| `server.address()` | `(...) => any` | `__tls_(ssl).server.address` | 📋 Planned | - |
| `server.close([callback])` | `(...) => any` | `__tls_(ssl).server.close` | 📋 Planned | - |
| `server.getTicketKeys()` | `(...) => any` | `__tls_(ssl).server.getTicketKeys` | 📋 Planned | - |
| `server.listen()` | `(...) => any` | `__tls_(ssl).server.listen` | 📋 Planned | - |
| `server.setSecureContext(options)` | `(...) => any` | `__tls_(ssl).server.setSecureContext` | 📋 Planned | - |
| `server.setTicketKeys(keys)` | `(...) => any` | `__tls_(ssl).server.setTicketKeys` | 📋 Planned | - |
| `tls.DEFAULT_ECDH_CURVE` | `any` | `__tls_(ssl).tls.DEFAULT_ECDH_CURVE` | 📋 Planned | - |
| `tls.SecurePair` | `(...) => any` | `__tls_(ssl).tls.SecurePair` | 📋 Planned | - |
| `tls.Server` | `(...) => any` | `__tls_(ssl).tls.Server` | 📋 Planned | - |
| `tls.TLSSocket` | `(...) => any` | `__tls_(ssl).tls.TLSSocket` | 📋 Planned | - |
| `tls.checkServerIdentity(hostname, cert)` | `(...) => any` | `__tls_(ssl).tls.checkServerIdentity` | 📋 Planned | - |
| `tls.connect(options[, callback])` | `(...) => any` | `__tls_(ssl).tls.connect` | 📋 Planned | - |
| `tls.connect(path[, options][, callback])` | `(...) => any` | `__tls_(ssl).tls.connect` | 📋 Planned | - |
| `tls.connect(port[, host][, options][, callback])` | `(...) => any` | `__tls_(ssl).tls.connect` | 📋 Planned | - |
| `tls.createSecureContext([options])` | `(...) => any` | `__tls_(ssl).tls.createSecureContext` | 📋 Planned | - |
| `tls.createSecurePair([context][, isServer][, requestCert][, rejectUnauthorized][, options])` | `(...) => any` | `__tls_(ssl).tls.createSecurePair` | 📋 Planned | - |
| `tls.createServer([options][, secureConnectionListener])` | `(...) => any` | `__tls_(ssl).tls.createServer` | 📋 Planned | - |
| `tls.getCACertificates([type])` | `(...) => any` | `__tls_(ssl).tls.getCACertificates` | 📋 Planned | - |
| `tls.getCiphers()` | `(...) => any` | `__tls_(ssl).tls.getCiphers` | 📋 Planned | - |
| `tls.setDefaultCACertificates(certs)` | `(...) => any` | `__tls_(ssl).tls.setDefaultCACertificates` | 📋 Planned | - |
| `tlsSocket.address()` | `(...) => any` | `__tls_(ssl).tlsSocket.address` | 📋 Planned | - |
| `tlsSocket.authorizationError` | `any` | `__tls_(ssl).tlsSocket.authorizationError` | 📋 Planned | - |
| `tlsSocket.disableRenegotiation()` | `(...) => any` | `__tls_(ssl).tlsSocket.disableRenegotiation` | 📋 Planned | - |
| `tlsSocket.enableTrace()` | `(...) => any` | `__tls_(ssl).tlsSocket.enableTrace` | 📋 Planned | - |
| `tlsSocket.encrypted` | `any` | `__tls_(ssl).tlsSocket.encrypted` | 📋 Planned | - |
| `tlsSocket.exportKeyingMaterial(length, label[, context])` | `(...) => any` | `__tls_(ssl).tlsSocket.exportKeyingMaterial` | 📋 Planned | - |
| `tlsSocket.getCertificate()` | `(...) => any` | `__tls_(ssl).tlsSocket.getCertificate` | 📋 Planned | - |
| `tlsSocket.getCipher()` | `(...) => any` | `__tls_(ssl).tlsSocket.getCipher` | 📋 Planned | - |
| `tlsSocket.getEphemeralKeyInfo()` | `(...) => any` | `__tls_(ssl).tlsSocket.getEphemeralKeyInfo` | 📋 Planned | - |
| `tlsSocket.getFinished()` | `(...) => any` | `__tls_(ssl).tlsSocket.getFinished` | 📋 Planned | - |
| `tlsSocket.getPeerCertificate([detailed])` | `(...) => any` | `__tls_(ssl).tlsSocket.getPeerCertificate` | 📋 Planned | - |
| `tlsSocket.getPeerFinished()` | `(...) => any` | `__tls_(ssl).tlsSocket.getPeerFinished` | 📋 Planned | - |
| `tlsSocket.getPeerX509Certificate()` | `(...) => any` | `__tls_(ssl).tlsSocket.getPeerX509Certificate` | 📋 Planned | - |
| `tlsSocket.getProtocol()` | `(...) => any` | `__tls_(ssl).tlsSocket.getProtocol` | 📋 Planned | - |
| `tlsSocket.getSession()` | `(...) => any` | `__tls_(ssl).tlsSocket.getSession` | 📋 Planned | - |
| `tlsSocket.getSharedSigalgs()` | `(...) => any` | `__tls_(ssl).tlsSocket.getSharedSigalgs` | 📋 Planned | - |
| `tlsSocket.getTLSTicket()` | `(...) => any` | `__tls_(ssl).tlsSocket.getTLSTicket` | 📋 Planned | - |
| `tlsSocket.getX509Certificate()` | `(...) => any` | `__tls_(ssl).tlsSocket.getX509Certificate` | 📋 Planned | - |
| `tlsSocket.isSessionReused()` | `(...) => any` | `__tls_(ssl).tlsSocket.isSessionReused` | 📋 Planned | - |
| `tlsSocket.renegotiate(options, callback)` | `(...) => any` | `__tls_(ssl).tlsSocket.renegotiate` | 📋 Planned | - |
| `tlsSocket.setKeyCert(context)` | `(...) => any` | `__tls_(ssl).tlsSocket.setKeyCert` | 📋 Planned | - |
| `tlsSocket.setMaxSendFragment(size)` | `(...) => any` | `__tls_(ssl).tlsSocket.setMaxSendFragment` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `tls_(ssl)` are organized per API under `internal/compiler/testdata/corpus/tls_(ssl)/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/tls_(ssl)/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
