# Util Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:util`  
> **Specification Reference**: [Node.js 22 LTS Util Documentation](https://nodejs.org/docs/latest-v22.x/api/util.html)  
> **Type Definition Source**: [@types/node/util.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-util-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:util`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `util.deprecate(fn, msg[, code])` | `(...) => any` | `__util.util.deprecate` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.format(format[, ...args])` | `(...) => any` | `__util.util.format` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.getSystemErrorMap()` | `(...) => any` | `__util.util.getSystemErrorMap` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.getSystemErrorName(err)` | `(...) => any` | `__util.util.getSystemErrorName` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.inspect(object[, options])` | `(...) => any` | `__util.util.inspect` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.inspect(object[, showHidden[, depth[, colors]]])` | `(...) => any` | `__util.util.inspect` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.isDeepStrictEqual(val1, val2)` | `(...) => any` | `__util.util.isDeepStrictEqual` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.parseEnv(content)` | `(...) => any` | `__util.util.parseEnv` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.stripVTControlCharacters(str)` | `(...) => any` | `__util.util.stripVTControlCharacters` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.styleText(format, text[, options])` | `(...) => any` | `__util.util.styleText` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.toUSVString(string)` | `(...) => any` | `__util.util.toUSVString` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `util.types` | `any` | `__util.util.types` | ✅ Done | `internal/compiler/testdata/corpus/api/util.ts` |
| `encoding` | `any` | `__util.encoding` | 📋 Planned | - |
| `essence` | `any` | `__util.essence` | 📋 Planned | - |
| `fatal` | `any` | `__util.fatal` | 📋 Planned | - |
| `ignoreBOM` | `any` | `__util.ignoreBOM` | 📋 Planned | - |
| `mime.toJSON()` | `(...) => any` | `__util.mime.toJSON` | 📋 Planned | - |
| `mime.toString()` | `(...) => any` | `__util.mime.toString` | 📋 Planned | - |
| `mimeParams.delete(name)` | `(...) => any` | `__util.mimeParams.delete` | 📋 Planned | - |
| `mimeParams.entries()` | `(...) => any` | `__util.mimeParams.entries` | 📋 Planned | - |
| `mimeParams.get(name)` | `(...) => any` | `__util.mimeParams.get` | 📋 Planned | - |
| `mimeParams.has(name)` | `(...) => any` | `__util.mimeParams.has` | 📋 Planned | - |
| `mimeParams.keys()` | `(...) => any` | `__util.mimeParams.keys` | 📋 Planned | - |
| `mimeParams.set(name, value)` | `(...) => any` | `__util.mimeParams.set` | 📋 Planned | - |
| `mimeParams.values()` | `(...) => any` | `__util.mimeParams.values` | 📋 Planned | - |
| `mimeParams[Symbol.iterator]()` | `(...) => any` | `__util.mimeParams[Symbol.iterator]` | 📋 Planned | - |
| `params` | `any` | `__util.params` | 📋 Planned | - |
| `subtype` | `any` | `__util.subtype` | 📋 Planned | - |
| `textDecoder.decode([input[, options]])` | `(...) => any` | `__util.textDecoder.decode` | 📋 Planned | - |
| `textEncoder.encode([input])` | `(...) => any` | `__util.textEncoder.encode` | 📋 Planned | - |
| `textEncoder.encodeInto(src, dest)` | `(...) => any` | `__util.textEncoder.encodeInto` | 📋 Planned | - |
| `type` | `any` | `__util.type` | 📋 Planned | - |
| `util.MIMEParams` | `(...) => any` | `__util.util.MIMEParams` | 📋 Planned | - |
| `util.MIMEType` | `(...) => any` | `__util.util.MIMEType` | 📋 Planned | - |
| `util.TextDecoder` | `(...) => any` | `__util.util.TextDecoder` | 📋 Planned | - |
| `util.TextEncoder` | `(...) => any` | `__util.util.TextEncoder` | 📋 Planned | - |
| `util._extend(target, source)` | `(...) => any` | `__util.util._extend` | 📋 Planned | - |
| `util.aborted(signal, resource)` | `(...) => any` | `__util.util.aborted` | 📋 Planned | - |
| `util.callbackify(original)` | `(...) => any` | `__util.util.callbackify` | 📋 Planned | - |
| `util.debug(section)` | `(...) => any` | `__util.util.debug` | 📋 Planned | - |
| `util.debuglog(section[, callback])` | `(...) => any` | `__util.util.debuglog` | 📋 Planned | - |
| `util.diff(actual, expected)` | `(...) => any` | `__util.util.diff` | 📋 Planned | - |
| `util.formatWithOptions(inspectOptions, format[, ...args])` | `(...) => any` | `__util.util.formatWithOptions` | 📋 Planned | - |
| `util.getCallSites([frameCount][, options])` | `(...) => any` | `__util.util.getCallSites` | 📋 Planned | - |
| `util.getSystemErrorMessage(err)` | `(...) => any` | `__util.util.getSystemErrorMessage` | 📋 Planned | - |
| `util.inherits(constructor, superConstructor)` | `(...) => any` | `__util.util.inherits` | 📋 Planned | - |
| `util.isArray(object)` | `(...) => any` | `__util.util.isArray` | 📋 Planned | - |
| `util.isBoolean(object)` | `(...) => any` | `__util.util.isBoolean` | 📋 Planned | - |
| `util.isBuffer(object)` | `(...) => any` | `__util.util.isBuffer` | 📋 Planned | - |
| `util.isDate(object)` | `(...) => any` | `__util.util.isDate` | 📋 Planned | - |
| `util.isError(object)` | `(...) => any` | `__util.util.isError` | 📋 Planned | - |
| `util.isFunction(object)` | `(...) => any` | `__util.util.isFunction` | 📋 Planned | - |
| `util.isNull(object)` | `(...) => any` | `__util.util.isNull` | 📋 Planned | - |
| `util.isNullOrUndefined(object)` | `(...) => any` | `__util.util.isNullOrUndefined` | 📋 Planned | - |
| `util.isNumber(object)` | `(...) => any` | `__util.util.isNumber` | 📋 Planned | - |
| `util.isObject(object)` | `(...) => any` | `__util.util.isObject` | 📋 Planned | - |
| `util.isPrimitive(object)` | `(...) => any` | `__util.util.isPrimitive` | 📋 Planned | - |
| `util.isRegExp(object)` | `(...) => any` | `__util.util.isRegExp` | 📋 Planned | - |
| `util.isString(object)` | `(...) => any` | `__util.util.isString` | 📋 Planned | - |
| `util.isSymbol(object)` | `(...) => any` | `__util.util.isSymbol` | 📋 Planned | - |
| `util.isUndefined(object)` | `(...) => any` | `__util.util.isUndefined` | 📋 Planned | - |
| `util.log(string)` | `(...) => any` | `__util.util.log` | 📋 Planned | - |
| `util.parseArgs([config])` | `(...) => any` | `__util.util.parseArgs` | 📋 Planned | - |
| `util.promisify(original)` | `(...) => any` | `__util.util.promisify` | 📋 Planned | - |
| `util.setTraceSigInt(enable)` | `(...) => any` | `__util.util.setTraceSigInt` | 📋 Planned | - |
| `util.transferableAbortController()` | `(...) => any` | `__util.util.transferableAbortController` | 📋 Planned | - |
| `util.transferableAbortSignal(signal)` | `(...) => any` | `__util.util.transferableAbortSignal` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `util` are organized per API under `internal/compiler/testdata/corpus/util/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/util/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
