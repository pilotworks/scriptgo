# URL Implementation Checklist

> **Category**: `CategoryECMAScript`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [TC39 ECMA-262 URL Specification](https://tc39.es/ecma262/#sec-url-objects)  
> **Type Definition Source**: [microsoft/TypeScript lib.es2024.d.ts](https://github.com/microsoft/TypeScript/tree/main/src/lib)  
> **Gate Oracle**: TC39 Test262 Test Suite & TypeScript baselines

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
| `URL.canParse(url: string, base?: string \| URL): boolean` | `canParse(url: string, base?: string \| URL): boolean` | `__url.canParse` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.hash: string` | `hash: string` | `__url.hash` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.host: string` | `host: string` | `__url.host` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.hostname: string` | `hostname: string` | `__url.hostname` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.href: string` | `href: string` | `__url.href` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.origin: string` | `origin: string` | `__url.origin` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.password: string` | `password: string` | `__url.password` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.pathname: string` | `pathname: string` | `__url.pathname` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.port: string` | `port: string` | `__url.port` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.protocol: string` | `protocol: string` | `__url.protocol` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.search: string` | `search: string` | `__url.search` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.searchParams: URLSearchParams` | `searchParams: URLSearchParams` | `__url.searchParams` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.toJSON(): string` | `toJSON(): string` | `__url.toJSON` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.toString(): string` | `toString(): string` | `__url.toString` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `URL.username: string` | `username: string` | `__url.username` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |
| `new URL(url: string, base?: string \| URL): URL` | `new(url: string, base?: string \| URL): URL` | `__url.new` | ✅ Done | `internal/compiler/testdata/corpus/api/url.ts` |

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
