# TTY Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:tty`  
> **Specification Reference**: [Node.js 22 LTS TTY Documentation](https://nodejs.org/docs/latest-v22.x/api/tty.html)  
> **Type Definition Source**: [@types/node/tty.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-tty-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:tty`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `readStream.isRaw` | `any` | `__tty.readStream.isRaw` | 📋 Planned | - |
| `readStream.isTTY` | `any` | `__tty.readStream.isTTY` | 📋 Planned | - |
| `readStream.setRawMode(mode)` | `(...) => any` | `__tty.readStream.setRawMode` | 📋 Planned | - |
| `tty.ReadStream` | `(...) => any` | `__tty.tty.ReadStream` | 📋 Planned | - |
| `tty.WriteStream` | `(...) => any` | `__tty.tty.WriteStream` | 📋 Planned | - |
| `tty.isatty(fd)` | `(...) => any` | `__tty.tty.isatty` | 📋 Planned | - |
| `writeStream.clearLine(dir[, callback])` | `(...) => any` | `__tty.writeStream.clearLine` | 📋 Planned | - |
| `writeStream.clearScreenDown([callback])` | `(...) => any` | `__tty.writeStream.clearScreenDown` | 📋 Planned | - |
| `writeStream.columns` | `any` | `__tty.writeStream.columns` | 📋 Planned | - |
| `writeStream.cursorTo(x[, y][, callback])` | `(...) => any` | `__tty.writeStream.cursorTo` | 📋 Planned | - |
| `writeStream.getColorDepth([env])` | `(...) => any` | `__tty.writeStream.getColorDepth` | 📋 Planned | - |
| `writeStream.getWindowSize()` | `(...) => any` | `__tty.writeStream.getWindowSize` | 📋 Planned | - |
| `writeStream.hasColors([count][, env])` | `(...) => any` | `__tty.writeStream.hasColors` | 📋 Planned | - |
| `writeStream.isTTY` | `any` | `__tty.writeStream.isTTY` | 📋 Planned | - |
| `writeStream.moveCursor(dx, dy[, callback])` | `(...) => any` | `__tty.writeStream.moveCursor` | 📋 Planned | - |
| `writeStream.rows` | `any` | `__tty.writeStream.rows` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `tty` are organized per API under `internal/compiler/testdata/corpus/tty/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 5: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/tty/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 6: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
