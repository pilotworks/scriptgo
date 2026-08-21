# URL Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS URL Documentation](https://nodejs.org/docs/latest-v22.x/api/url.html)  
> **Type Definition Source**: [@types/node/url.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-url-*.js)

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
| `URL` | `(...) => any` | `__url.URL` | ✅ Done | `internal/compiler/testdata/corpus/api/url/URL/` |
| `URLSearchParams` | `(...) => any` | `__url.URLSearchParams` | ✅ Done | `internal/compiler/testdata/corpus/api/url/URLSearchParams/` |
| `URL.canParse(input[, base])` | `(...) => any` | `__url.URL.canParse` | 📋 Planned | - |
| `URL.createObjectURL(blob)` | `(...) => any` | `__url.URL.createObjectURL` | 📋 Planned | - |
| `URL.parse(input[, base])` | `(...) => any` | `__url.URL.parse` | 📋 Planned | - |
| `URL.revokeObjectURL(id)` | `(...) => any` | `__url.URL.revokeObjectURL` | 📋 Planned | - |
| `hash` | `any` | `__url.hash` | 📋 Planned | - |
| `host` | `any` | `__url.host` | 📋 Planned | - |
| `hostname` | `any` | `__url.hostname` | 📋 Planned | - |
| `href` | `any` | `__url.href` | 📋 Planned | - |
| `origin` | `any` | `__url.origin` | 📋 Planned | - |
| `password` | `any` | `__url.password` | 📋 Planned | - |
| `pathname` | `any` | `__url.pathname` | 📋 Planned | - |
| `port` | `any` | `__url.port` | 📋 Planned | - |
| `protocol` | `any` | `__url.protocol` | 📋 Planned | - |
| `search` | `any` | `__url.search` | 📋 Planned | - |
| `searchParams` | `any` | `__url.searchParams` | 📋 Planned | - |
| `url.domainToASCII(domain)` | `(...) => any` | `__url.url.domainToASCII` | 📋 Planned | - |
| `url.domainToUnicode(domain)` | `(...) => any` | `__url.url.domainToUnicode` | 📋 Planned | - |
| `url.fileURLToPath(url[, options])` | `(...) => any` | `__url.url.fileURLToPath` | 📋 Planned | - |
| `url.fileURLToPathBuffer(url[, options])` | `(...) => any` | `__url.url.fileURLToPathBuffer` | 📋 Planned | - |
| `url.format(URL[, options])` | `(...) => any` | `__url.url.format` | 📋 Planned | - |
| `url.format(urlObject)` | `(...) => any` | `__url.url.format` | 📋 Planned | - |
| `url.parse(urlString[, parseQueryString[, slashesDenoteHost]])` | `(...) => any` | `__url.url.parse` | 📋 Planned | - |
| `url.pathToFileURL(path[, options])` | `(...) => any` | `__url.url.pathToFileURL` | 📋 Planned | - |
| `url.resolve(from, to)` | `(...) => any` | `__url.url.resolve` | 📋 Planned | - |
| `url.toJSON()` | `(...) => any` | `__url.url.toJSON` | 📋 Planned | - |
| `url.toString()` | `(...) => any` | `__url.url.toString` | 📋 Planned | - |
| `url.urlToHttpOptions(url)` | `(...) => any` | `__url.url.urlToHttpOptions` | 📋 Planned | - |
| `urlSearchParams.append(name, value)` | `(...) => any` | `__url.urlSearchParams.append` | 📋 Planned | - |
| `urlSearchParams.delete(name[, value])` | `(...) => any` | `__url.urlSearchParams.delete` | 📋 Planned | - |
| `urlSearchParams.entries()` | `(...) => any` | `__url.urlSearchParams.entries` | 📋 Planned | - |
| `urlSearchParams.forEach(fn[, thisArg])` | `(...) => any` | `__url.urlSearchParams.forEach` | 📋 Planned | - |
| `urlSearchParams.get(name)` | `(...) => any` | `__url.urlSearchParams.get` | 📋 Planned | - |
| `urlSearchParams.getAll(name)` | `(...) => any` | `__url.urlSearchParams.getAll` | 📋 Planned | - |
| `urlSearchParams.has(name[, value])` | `(...) => any` | `__url.urlSearchParams.has` | 📋 Planned | - |
| `urlSearchParams.keys()` | `(...) => any` | `__url.urlSearchParams.keys` | 📋 Planned | - |
| `urlSearchParams.set(name, value)` | `(...) => any` | `__url.urlSearchParams.set` | 📋 Planned | - |
| `urlSearchParams.size` | `any` | `__url.urlSearchParams.size` | 📋 Planned | - |
| `urlSearchParams.sort()` | `(...) => any` | `__url.urlSearchParams.sort` | 📋 Planned | - |
| `urlSearchParams.toString()` | `(...) => any` | `__url.urlSearchParams.toString` | 📋 Planned | - |
| `urlSearchParams.values()` | `(...) => any` | `__url.urlSearchParams.values` | 📋 Planned | - |
| `urlSearchParams[Symbol.iterator]()` | `(...) => any` | `__url.urlSearchParams[Symbol.iterator]` | 📋 Planned | - |
| `username` | `any` | `__url.username` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `url` are organized per API under `internal/compiler/testdata/corpus/url/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/url/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
