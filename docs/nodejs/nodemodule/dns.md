# DNS Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:dns`  
> **Specification Reference**: [Node.js 22 LTS DNS Documentation](https://nodejs.org/docs/latest-v22.x/api/dns.html)  
> **Type Definition Source**: [@types/node/dns.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-dns-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:dns`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `Resolver([options])` | `(...) => any` | `__dns.Resolver` | 📋 Planned | - |
| `dns.Resolver` | `(...) => any` | `__dns.dns.Resolver` | 📋 Planned | - |
| `dns.getDefaultResultOrder()` | `(...) => any` | `__dns.dns.getDefaultResultOrder` | 📋 Planned | - |
| `dns.getServers()` | `(...) => any` | `__dns.dns.getServers` | 📋 Planned | - |
| `dns.lookup()` | `(...) => any` | `__dns.dns.lookup` | 📋 Planned | - |
| `dns.lookup(hostname[, options], callback)` | `(...) => any` | `__dns.dns.lookup` | 📋 Planned | - |
| `dns.lookupService(address, port, callback)` | `(...) => any` | `__dns.dns.lookupService` | 📋 Planned | - |
| `dns.resolve(hostname[, rrtype], callback)` | `(...) => any` | `__dns.dns.resolve` | 📋 Planned | - |
| `dns.resolve4(hostname[, options], callback)` | `(...) => any` | `__dns.dns.resolve4` | 📋 Planned | - |
| `dns.resolve6(hostname[, options], callback)` | `(...) => any` | `__dns.dns.resolve6` | 📋 Planned | - |
| `dns.resolveAny(hostname, callback)` | `(...) => any` | `__dns.dns.resolveAny` | 📋 Planned | - |
| `dns.resolveCaa(hostname, callback)` | `(...) => any` | `__dns.dns.resolveCaa` | 📋 Planned | - |
| `dns.resolveCname(hostname, callback)` | `(...) => any` | `__dns.dns.resolveCname` | 📋 Planned | - |
| `dns.resolveMx(hostname, callback)` | `(...) => any` | `__dns.dns.resolveMx` | 📋 Planned | - |
| `dns.resolveNaptr(hostname, callback)` | `(...) => any` | `__dns.dns.resolveNaptr` | 📋 Planned | - |
| `dns.resolveNs(hostname, callback)` | `(...) => any` | `__dns.dns.resolveNs` | 📋 Planned | - |
| `dns.resolvePtr(hostname, callback)` | `(...) => any` | `__dns.dns.resolvePtr` | 📋 Planned | - |
| `dns.resolveSoa(hostname, callback)` | `(...) => any` | `__dns.dns.resolveSoa` | 📋 Planned | - |
| `dns.resolveSrv(hostname, callback)` | `(...) => any` | `__dns.dns.resolveSrv` | 📋 Planned | - |
| `dns.resolveTlsa(hostname, callback)` | `(...) => any` | `__dns.dns.resolveTlsa` | 📋 Planned | - |
| `dns.resolveTxt(hostname, callback)` | `(...) => any` | `__dns.dns.resolveTxt` | 📋 Planned | - |
| `dns.reverse(ip, callback)` | `(...) => any` | `__dns.dns.reverse` | 📋 Planned | - |
| `dns.setDefaultResultOrder(order)` | `(...) => any` | `__dns.dns.setDefaultResultOrder` | 📋 Planned | - |
| `dns.setServers(servers)` | `(...) => any` | `__dns.dns.setServers` | 📋 Planned | - |
| `dnsPromises.Resolver` | `(...) => any` | `__dns.dnsPromises.Resolver` | 📋 Planned | - |
| `dnsPromises.getDefaultResultOrder()` | `(...) => any` | `__dns.dnsPromises.getDefaultResultOrder` | 📋 Planned | - |
| `dnsPromises.getServers()` | `(...) => any` | `__dns.dnsPromises.getServers` | 📋 Planned | - |
| `dnsPromises.lookup(hostname[, options])` | `(...) => any` | `__dns.dnsPromises.lookup` | 📋 Planned | - |
| `dnsPromises.lookupService(address, port)` | `(...) => any` | `__dns.dnsPromises.lookupService` | 📋 Planned | - |
| `dnsPromises.resolve(hostname[, rrtype])` | `(...) => any` | `__dns.dnsPromises.resolve` | 📋 Planned | - |
| `dnsPromises.resolve4(hostname[, options])` | `(...) => any` | `__dns.dnsPromises.resolve4` | 📋 Planned | - |
| `dnsPromises.resolve6(hostname[, options])` | `(...) => any` | `__dns.dnsPromises.resolve6` | 📋 Planned | - |
| `dnsPromises.resolveAny(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveAny` | 📋 Planned | - |
| `dnsPromises.resolveCaa(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveCaa` | 📋 Planned | - |
| `dnsPromises.resolveCname(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveCname` | 📋 Planned | - |
| `dnsPromises.resolveMx(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveMx` | 📋 Planned | - |
| `dnsPromises.resolveNaptr(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveNaptr` | 📋 Planned | - |
| `dnsPromises.resolveNs(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveNs` | 📋 Planned | - |
| `dnsPromises.resolvePtr(hostname)` | `(...) => any` | `__dns.dnsPromises.resolvePtr` | 📋 Planned | - |
| `dnsPromises.resolveSoa(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveSoa` | 📋 Planned | - |
| `dnsPromises.resolveSrv(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveSrv` | 📋 Planned | - |
| `dnsPromises.resolveTlsa(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveTlsa` | 📋 Planned | - |
| `dnsPromises.resolveTxt(hostname)` | `(...) => any` | `__dns.dnsPromises.resolveTxt` | 📋 Planned | - |
| `dnsPromises.reverse(ip)` | `(...) => any` | `__dns.dnsPromises.reverse` | 📋 Planned | - |
| `dnsPromises.setDefaultResultOrder(order)` | `(...) => any` | `__dns.dnsPromises.setDefaultResultOrder` | 📋 Planned | - |
| `dnsPromises.setServers(servers)` | `(...) => any` | `__dns.dnsPromises.setServers` | 📋 Planned | - |
| `resolver.cancel()` | `(...) => any` | `__dns.resolver.cancel` | 📋 Planned | - |
| `resolver.setLocalAddress([ipv4][, ipv6])` | `(...) => any` | `__dns.resolver.setLocalAddress` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `dns` are organized per API under `internal/compiler/testdata/corpus/dns/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/dns/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
