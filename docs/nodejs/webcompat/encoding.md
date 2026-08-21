# TextEncoder & TextDecoder API Implementation Checklist

> **Category**: `CategoryWebCompat`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [WinterCG / WHATWG TextEncoder & TextDecoder API Specification](https://wintercg.org/)  
> **Type Definition Source**: [microsoft/TypeScript lib.dom.d.ts (Server subset)](https://github.com/microsoft/TypeScript/blob/main/src/lib/lib.dom.d.ts)  
> **Gate Oracle**: Web Platform Tests (WPT) & Node.js 22 LTS WPT runner

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
| `encoding` | `any` | `__encoding.encoding` | 📋 Planned | - |
| `essence` | `any` | `__encoding.essence` | 📋 Planned | - |
| `fatal` | `any` | `__encoding.fatal` | 📋 Planned | - |
| `ignoreBOM` | `any` | `__encoding.ignoreBOM` | 📋 Planned | - |
| `mime.toJSON()` | `(...) => any` | `__encoding.mime.toJSON` | 📋 Planned | - |
| `mime.toString()` | `(...) => any` | `__encoding.mime.toString` | 📋 Planned | - |
| `mimeParams.delete(name)` | `(...) => any` | `__encoding.mimeParams.delete` | 📋 Planned | - |
| `mimeParams.entries()` | `(...) => any` | `__encoding.mimeParams.entries` | 📋 Planned | - |
| `mimeParams.get(name)` | `(...) => any` | `__encoding.mimeParams.get` | 📋 Planned | - |
| `mimeParams.has(name)` | `(...) => any` | `__encoding.mimeParams.has` | 📋 Planned | - |
| `mimeParams.keys()` | `(...) => any` | `__encoding.mimeParams.keys` | 📋 Planned | - |
| `mimeParams.set(name, value)` | `(...) => any` | `__encoding.mimeParams.set` | 📋 Planned | - |
| `mimeParams.values()` | `(...) => any` | `__encoding.mimeParams.values` | 📋 Planned | - |
| `mimeParams[Symbol.iterator]()` | `(...) => any` | `__encoding.mimeParams[Symbol.iterator]` | 📋 Planned | - |
| `params` | `any` | `__encoding.params` | 📋 Planned | - |
| `subtype` | `any` | `__encoding.subtype` | 📋 Planned | - |
| `textDecoder.decode([input[, options]])` | `(...) => any` | `__encoding.textDecoder.decode` | 📋 Planned | - |
| `textEncoder.encode([input])` | `(...) => any` | `__encoding.textEncoder.encode` | 📋 Planned | - |
| `textEncoder.encodeInto(src, dest)` | `(...) => any` | `__encoding.textEncoder.encodeInto` | 📋 Planned | - |
| `type` | `any` | `__encoding.type` | 📋 Planned | - |
| `util.MIMEParams` | `(...) => any` | `__encoding.util.MIMEParams` | 📋 Planned | - |
| `util.MIMEType` | `(...) => any` | `__encoding.util.MIMEType` | 📋 Planned | - |
| `util.TextDecoder` | `(...) => any` | `__encoding.util.TextDecoder` | 📋 Planned | - |
| `util.TextEncoder` | `(...) => any` | `__encoding.util.TextEncoder` | 📋 Planned | - |
| `util._extend(target, source)` | `(...) => any` | `__encoding.util._extend` | 📋 Planned | - |
| `util.aborted(signal, resource)` | `(...) => any` | `__encoding.util.aborted` | 📋 Planned | - |
| `util.callbackify(original)` | `(...) => any` | `__encoding.util.callbackify` | 📋 Planned | - |
| `util.debug(section)` | `(...) => any` | `__encoding.util.debug` | 📋 Planned | - |
| `util.debuglog(section[, callback])` | `(...) => any` | `__encoding.util.debuglog` | 📋 Planned | - |
| `util.deprecate(fn, msg[, code])` | `(...) => any` | `__encoding.util.deprecate` | 📋 Planned | - |
| `util.diff(actual, expected)` | `(...) => any` | `__encoding.util.diff` | 📋 Planned | - |
| `util.format(format[, ...args])` | `(...) => any` | `__encoding.util.format` | 📋 Planned | - |
| `util.formatWithOptions(inspectOptions, format[, ...args])` | `(...) => any` | `__encoding.util.formatWithOptions` | 📋 Planned | - |
| `util.getCallSites([frameCount][, options])` | `(...) => any` | `__encoding.util.getCallSites` | 📋 Planned | - |
| `util.getSystemErrorMap()` | `(...) => any` | `__encoding.util.getSystemErrorMap` | 📋 Planned | - |
| `util.getSystemErrorMessage(err)` | `(...) => any` | `__encoding.util.getSystemErrorMessage` | 📋 Planned | - |
| `util.getSystemErrorName(err)` | `(...) => any` | `__encoding.util.getSystemErrorName` | 📋 Planned | - |
| `util.inherits(constructor, superConstructor)` | `(...) => any` | `__encoding.util.inherits` | 📋 Planned | - |
| `util.inspect(object[, options])` | `(...) => any` | `__encoding.util.inspect` | 📋 Planned | - |
| `util.inspect(object[, showHidden[, depth[, colors]]])` | `(...) => any` | `__encoding.util.inspect` | 📋 Planned | - |
| `util.isArray(object)` | `(...) => any` | `__encoding.util.isArray` | 📋 Planned | - |
| `util.isBoolean(object)` | `(...) => any` | `__encoding.util.isBoolean` | 📋 Planned | - |
| `util.isBuffer(object)` | `(...) => any` | `__encoding.util.isBuffer` | 📋 Planned | - |
| `util.isDate(object)` | `(...) => any` | `__encoding.util.isDate` | 📋 Planned | - |
| `util.isDeepStrictEqual(val1, val2)` | `(...) => any` | `__encoding.util.isDeepStrictEqual` | 📋 Planned | - |
| `util.isError(object)` | `(...) => any` | `__encoding.util.isError` | 📋 Planned | - |
| `util.isFunction(object)` | `(...) => any` | `__encoding.util.isFunction` | 📋 Planned | - |
| `util.isNull(object)` | `(...) => any` | `__encoding.util.isNull` | 📋 Planned | - |
| `util.isNullOrUndefined(object)` | `(...) => any` | `__encoding.util.isNullOrUndefined` | 📋 Planned | - |
| `util.isNumber(object)` | `(...) => any` | `__encoding.util.isNumber` | 📋 Planned | - |
| `util.isObject(object)` | `(...) => any` | `__encoding.util.isObject` | 📋 Planned | - |
| `util.isPrimitive(object)` | `(...) => any` | `__encoding.util.isPrimitive` | 📋 Planned | - |
| `util.isRegExp(object)` | `(...) => any` | `__encoding.util.isRegExp` | 📋 Planned | - |
| `util.isString(object)` | `(...) => any` | `__encoding.util.isString` | 📋 Planned | - |
| `util.isSymbol(object)` | `(...) => any` | `__encoding.util.isSymbol` | 📋 Planned | - |
| `util.isUndefined(object)` | `(...) => any` | `__encoding.util.isUndefined` | 📋 Planned | - |
| `util.log(string)` | `(...) => any` | `__encoding.util.log` | 📋 Planned | - |
| `util.parseArgs([config])` | `(...) => any` | `__encoding.util.parseArgs` | 📋 Planned | - |
| `util.parseEnv(content)` | `(...) => any` | `__encoding.util.parseEnv` | 📋 Planned | - |
| `util.promisify(original)` | `(...) => any` | `__encoding.util.promisify` | 📋 Planned | - |
| `util.setTraceSigInt(enable)` | `(...) => any` | `__encoding.util.setTraceSigInt` | 📋 Planned | - |
| `util.stripVTControlCharacters(str)` | `(...) => any` | `__encoding.util.stripVTControlCharacters` | 📋 Planned | - |
| `util.styleText(format, text[, options])` | `(...) => any` | `__encoding.util.styleText` | 📋 Planned | - |
| `util.toUSVString(string)` | `(...) => any` | `__encoding.util.toUSVString` | 📋 Planned | - |
| `util.transferableAbortController()` | `(...) => any` | `__encoding.util.transferableAbortController` | 📋 Planned | - |
| `util.transferableAbortSignal(signal)` | `(...) => any` | `__encoding.util.transferableAbortSignal` | 📋 Planned | - |
| `util.types` | `any` | `__encoding.util.types` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `encoding` are organized per API under `internal/compiler/testdata/corpus/encoding/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/encoding/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
