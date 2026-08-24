# URL Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:url`  
> **Specification Reference**: [Node.js 22 LTS URL Documentation](https://nodejs.org/docs/latest-v22.x/api/url.html)  
> **Type Definition Source**: [@types/node/url.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-url-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:url`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `URL` | `(...) => any` | `__url.URL` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.canParse(input[, base])` | `(...) => any` | `__url.URL.canParse` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.createObjectURL(blob)` | `(...) => any` | `__url.URL.createObjectURL` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.parse(input[, base])` | `(...) => any` | `__url.URL.parse` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.revokeObjectURL(id)` | `(...) => any` | `__url.URL.revokeObjectURL` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URLSearchParams` | `(...) => any` | `__url.URLSearchParams` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `hash` | `any` | `__url.hash` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `host` | `any` | `__url.host` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `hostname` | `any` | `__url.hostname` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `href` | `any` | `__url.href` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `origin` | `any` | `__url.origin` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `password` | `any` | `__url.password` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `pathname` | `any` | `__url.pathname` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `port` | `any` | `__url.port` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `protocol` | `any` | `__url.protocol` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `search` | `any` | `__url.search` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `searchParams` | `any` | `__url.searchParams` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.domainToASCII(domain)` | `(...) => any` | `__url.url.domainToASCII` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.domainToUnicode(domain)` | `(...) => any` | `__url.url.domainToUnicode` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.fileURLToPath(url[, options])` | `(...) => any` | `__url.url.fileURLToPath` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.fileURLToPathBuffer(url[, options])` | `(...) => any` | `__url.url.fileURLToPathBuffer` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.format(URL[, options])` | `(...) => any` | `__url.url.format` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.format(urlObject)` | `(...) => any` | `__url.url.format` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.parse(urlString[, parseQueryString[, slashesDenoteHost]])` | `(...) => any` | `__url.url.parse` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.pathToFileURL(path[, options])` | `(...) => any` | `__url.url.pathToFileURL` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.resolve(from, to)` | `(...) => any` | `__url.url.resolve` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.toJSON()` | `(...) => any` | `__url.url.toJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.toString()` | `(...) => any` | `__url.url.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `url.urlToHttpOptions(url)` | `(...) => any` | `__url.url.urlToHttpOptions` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.append(name, value)` | `(...) => any` | `__url.urlSearchParams.append` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.delete(name[, value])` | `(...) => any` | `__url.urlSearchParams.delete` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.entries()` | `(...) => any` | `__url.urlSearchParams.entries` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.forEach(fn[, thisArg])` | `(...) => any` | `__url.urlSearchParams.forEach` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.get(name)` | `(...) => any` | `__url.urlSearchParams.get` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.getAll(name)` | `(...) => any` | `__url.urlSearchParams.getAll` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.has(name[, value])` | `(...) => any` | `__url.urlSearchParams.has` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.keys()` | `(...) => any` | `__url.urlSearchParams.keys` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.set(name, value)` | `(...) => any` | `__url.urlSearchParams.set` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.size` | `any` | `__url.urlSearchParams.size` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.sort()` | `(...) => any` | `__url.urlSearchParams.sort` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.toString()` | `(...) => any` | `__url.urlSearchParams.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams.values()` | `(...) => any` | `__url.urlSearchParams.values` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `urlSearchParams[Symbol.iterator]()` | `(...) => any` | `__url.urlSearchParams[Symbol.iterator]` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `username` | `any` | `__url.username` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |

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
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/url/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
