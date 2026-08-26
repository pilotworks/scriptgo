# process Implementation Checklist

> **Category**: `CategoryNodeGlobal`  
> **Import Path**: `N/A (Global Scope)`  
> **Specification Reference**: [Node.js 22 LTS process Global Documentation](https://nodejs.org/docs/latest-v22.x/api/process.html)  
> **Type Definition Source**: [@types/node/process.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-process-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Auto-global ambient identifiers available in root execution scope without explicit imports.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `argv` | `any` | `__process.argv` | ✅ Done | `internal/compiler/testdata/corpus/api/process.ts` |
| `env` | `any` | `__process.env` | ✅ Done | `internal/compiler/testdata/corpus/api/process.ts` |
| `process.cwd()` | `(...) => any` | `__process.process.cwd` | ✅ Done | `internal/compiler/testdata/corpus/api/process.ts` |
| `process.exit([code])` | `(...) => any` | `__process.process.exit` | ✅ Done | `internal/compiler/testdata/corpus/api/process.ts` |
| `allowedNodeEnvironmentFlags` | `any` | `__process.allowedNodeEnvironmentFlags` | 📋 Planned | - |
| `arch` | `any` | `__process.arch` | 📋 Planned | - |
| `argv0` | `any` | `__process.argv0` | 📋 Planned | - |
| `cached_builtins` | `any` | `__process.cached_builtins` | 📋 Planned | - |
| `channel` | `any` | `__process.channel` | 📋 Planned | - |
| `config` | `any` | `__process.config` | 📋 Planned | - |
| `connected` | `any` | `__process.connected` | 📋 Planned | - |
| `debug` | `any` | `__process.debug` | 📋 Planned | - |
| `debugPort` | `any` | `__process.debugPort` | 📋 Planned | - |
| `execArgv` | `any` | `__process.execArgv` | 📋 Planned | - |
| `execPath` | `any` | `__process.execPath` | 📋 Planned | - |
| `exitCode` | `any` | `__process.exitCode` | 📋 Planned | - |
| `inspector` | `any` | `__process.inspector` | 📋 Planned | - |
| `ipv6` | `any` | `__process.ipv6` | 📋 Planned | - |
| `mainModule` | `any` | `__process.mainModule` | 📋 Planned | - |
| `noDeprecation` | `any` | `__process.noDeprecation` | 📋 Planned | - |
| `permission` | `any` | `__process.permission` | 📋 Planned | - |
| `pid` | `any` | `__process.pid` | 📋 Planned | - |
| `platform` | `any` | `__process.platform` | 📋 Planned | - |
| `ppid` | `any` | `__process.ppid` | 📋 Planned | - |
| `process.abort()` | `(...) => any` | `__process.process.abort` | 📋 Planned | - |
| `process.availableMemory()` | `(...) => any` | `__process.process.availableMemory` | 📋 Planned | - |
| `process.chdir(directory)` | `(...) => any` | `__process.process.chdir` | 📋 Planned | - |
| `process.constrainedMemory()` | `(...) => any` | `__process.process.constrainedMemory` | 📋 Planned | - |
| `process.cpuUsage([previousValue])` | `(...) => any` | `__process.process.cpuUsage` | 📋 Planned | - |
| `process.disconnect()` | `(...) => any` | `__process.process.disconnect` | 📋 Planned | - |
| `process.dlopen(module, filename[, flags])` | `(...) => any` | `__process.process.dlopen` | 📋 Planned | - |
| `process.emitWarning(warning[, options])` | `(...) => any` | `__process.process.emitWarning` | 📋 Planned | - |
| `process.emitWarning(warning[, type[, code]][, ctor])` | `(...) => any` | `__process.process.emitWarning` | 📋 Planned | - |
| `process.execve(file[, args[, env]])` | `(...) => any` | `__process.process.execve` | 📋 Planned | - |
| `process.finalization.register(ref, callback)` | `(...) => any` | `__process.process.finalization.register` | 📋 Planned | - |
| `process.finalization.registerBeforeExit(ref, callback)` | `(...) => any` | `__process.process.finalization.registerBeforeExit` | 📋 Planned | - |
| `process.finalization.unregister(ref)` | `(...) => any` | `__process.process.finalization.unregister` | 📋 Planned | - |
| `process.getActiveResourcesInfo()` | `(...) => any` | `__process.process.getActiveResourcesInfo` | 📋 Planned | - |
| `process.getBuiltinModule(id)` | `(...) => any` | `__process.process.getBuiltinModule` | 📋 Planned | - |
| `process.getegid()` | `(...) => any` | `__process.process.getegid` | 📋 Planned | - |
| `process.geteuid()` | `(...) => any` | `__process.process.geteuid` | 📋 Planned | - |
| `process.getgid()` | `(...) => any` | `__process.process.getgid` | 📋 Planned | - |
| `process.getgroups()` | `(...) => any` | `__process.process.getgroups` | 📋 Planned | - |
| `process.getuid()` | `(...) => any` | `__process.process.getuid` | 📋 Planned | - |
| `process.hasUncaughtExceptionCaptureCallback()` | `(...) => any` | `__process.process.hasUncaughtExceptionCaptureCallback` | 📋 Planned | - |
| `process.hrtime([time])` | `(...) => any` | `__process.process.hrtime` | 📋 Planned | - |
| `process.hrtime.bigint()` | `(...) => any` | `__process.process.hrtime.bigint` | 📋 Planned | - |
| `process.initgroups(user, extraGroup)` | `(...) => any` | `__process.process.initgroups` | 📋 Planned | - |
| `process.kill(pid[, signal])` | `(...) => any` | `__process.process.kill` | 📋 Planned | - |
| `process.loadEnvFile(path)` | `(...) => any` | `__process.process.loadEnvFile` | 📋 Planned | - |
| `process.memoryUsage()` | `(...) => any` | `__process.process.memoryUsage` | 📋 Planned | - |
| `process.memoryUsage.rss()` | `(...) => any` | `__process.process.memoryUsage.rss` | 📋 Planned | - |
| `process.nextTick(callback[, ...args])` | `(...) => any` | `__process.process.nextTick` | 📋 Planned | - |
| `process.ref(maybeRefable)` | `(...) => any` | `__process.process.ref` | 📋 Planned | - |
| `process.resourceUsage()` | `(...) => any` | `__process.process.resourceUsage` | 📋 Planned | - |
| `process.send(message[, sendHandle[, options]][, callback])` | `(...) => any` | `__process.process.send` | 📋 Planned | - |
| `process.setSourceMapsEnabled(val)` | `(...) => any` | `__process.process.setSourceMapsEnabled` | 📋 Planned | - |
| `process.setUncaughtExceptionCaptureCallback(fn)` | `(...) => any` | `__process.process.setUncaughtExceptionCaptureCallback` | 📋 Planned | - |
| `process.setegid(id)` | `(...) => any` | `__process.process.setegid` | 📋 Planned | - |
| `process.seteuid(id)` | `(...) => any` | `__process.process.seteuid` | 📋 Planned | - |
| `process.setgid(id)` | `(...) => any` | `__process.process.setgid` | 📋 Planned | - |
| `process.setgroups(groups)` | `(...) => any` | `__process.process.setgroups` | 📋 Planned | - |
| `process.setuid(id)` | `(...) => any` | `__process.process.setuid` | 📋 Planned | - |
| `process.threadCpuUsage([previousValue])` | `(...) => any` | `__process.process.threadCpuUsage` | 📋 Planned | - |
| `process.umask()` | `(...) => any` | `__process.process.umask` | 📋 Planned | - |
| `process.umask(mask)` | `(...) => any` | `__process.process.umask` | 📋 Planned | - |
| `process.unref(maybeRefable)` | `(...) => any` | `__process.process.unref` | 📋 Planned | - |
| `process.uptime()` | `(...) => any` | `__process.process.uptime` | 📋 Planned | - |
| `release` | `any` | `__process.release` | 📋 Planned | - |
| `report` | `any` | `__process.report` | 📋 Planned | - |
| `require_module` | `any` | `__process.require_module` | 📋 Planned | - |
| `sourceMapsEnabled` | `any` | `__process.sourceMapsEnabled` | 📋 Planned | - |
| `stderr` | `any` | `__process.stderr` | 📋 Planned | - |
| `stdin` | `any` | `__process.stdin` | 📋 Planned | - |
| `stdout` | `any` | `__process.stdout` | 📋 Planned | - |
| `throwDeprecation` | `any` | `__process.throwDeprecation` | 📋 Planned | - |
| `title` | `any` | `__process.title` | 📋 Planned | - |
| `tls` | `any` | `__process.tls` | 📋 Planned | - |
| `tls_alpn` | `any` | `__process.tls_alpn` | 📋 Planned | - |
| `tls_ocsp` | `any` | `__process.tls_ocsp` | 📋 Planned | - |
| `tls_sni` | `any` | `__process.tls_sni` | 📋 Planned | - |
| `traceDeprecation` | `any` | `__process.traceDeprecation` | 📋 Planned | - |
| `traceProcessWarnings` {boolean}` | `any` | `__process.traceProcessWarnings` {boolean}` | 📋 Planned | - |
| `typescript` | `any` | `__process.typescript` | 📋 Planned | - |
| `uv` | `any` | `__process.uv` | 📋 Planned | - |
| `version` | `any` | `__process.version` | 📋 Planned | - |
| `versions` | `any` | `__process.versions` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `process` are organized per API under `internal/compiler/testdata/corpus/process/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/process/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
