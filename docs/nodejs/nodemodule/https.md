# HTTPS Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:https`  
> **Specification Reference**: [Node.js 22 LTS HTTPS Documentation](https://nodejs.org/docs/latest-v22.x/api/https.html)  
> **Type Definition Source**: [@types/node/https.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-https-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:https`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `headersTimeout` | `any` | `__https.headersTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.Agent` | `(...) => any` | `__https.https.Agent` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.Server` | `(...) => any` | `__https.https.Server` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.createServer([options][, requestListener])` | `(...) => any` | `__https.https.createServer` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.get(options[, callback])` | `(...) => any` | `__https.https.get` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.get(url[, options][, callback])` | `(...) => any` | `__https.https.get` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.globalAgent` | `any` | `__https.https.globalAgent` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.request(options[, callback])` | `(...) => any` | `__https.https.request` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `https.request(url[, options][, callback])` | `(...) => any` | `__https.https.request` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `keepAliveTimeout` | `any` | `__https.keepAliveTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `maxHeadersCount` | `any` | `__https.maxHeadersCount` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `requestTimeout` | `any` | `__https.requestTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server.close([callback])` | `(...) => any` | `__https.server.close` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server.closeAllConnections()` | `(...) => any` | `__https.server.closeAllConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server.closeIdleConnections()` | `(...) => any` | `__https.server.closeIdleConnections` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server.listen()` | `(...) => any` | `__https.server.listen` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server.setTimeout([msecs][, callback])` | `(...) => any` | `__https.server.setTimeout` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `server[Symbol.asyncDispose]()` | `(...) => any` | `__https.server[Symbol.asyncDispose]` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |
| `timeout` | `any` | `__https.timeout` | ✅ Done | `internal/compiler/testdata/corpus/api/https.ts` |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `https` are organized per API under `internal/compiler/testdata/corpus/https/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/https/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
