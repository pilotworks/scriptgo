# Readline Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:readline`  
> **Specification Reference**: [Node.js 22 LTS Readline Documentation](https://nodejs.org/docs/latest-v22.x/api/readline.html)  
> **Type Definition Source**: [@types/node/readline.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-readline-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:readline`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `InterfaceConstructor` | `(...) => any` | `__readline.InterfaceConstructor` | 📋 Planned | - |
| `cursor` | `any` | `__readline.cursor` | 📋 Planned | - |
| `line` | `any` | `__readline.line` | 📋 Planned | - |
| `readline.Interface` | `(...) => any` | `__readline.readline.Interface` | 📋 Planned | - |
| `readline.clearLine(stream, dir[, callback])` | `(...) => any` | `__readline.readline.clearLine` | 📋 Planned | - |
| `readline.clearScreenDown(stream[, callback])` | `(...) => any` | `__readline.readline.clearScreenDown` | 📋 Planned | - |
| `readline.createInterface(options)` | `(...) => any` | `__readline.readline.createInterface` | 📋 Planned | - |
| `readline.cursorTo(stream, x[, y][, callback])` | `(...) => any` | `__readline.readline.cursorTo` | 📋 Planned | - |
| `readline.emitKeypressEvents(stream[, interface])` | `(...) => any` | `__readline.readline.emitKeypressEvents` | 📋 Planned | - |
| `readline.moveCursor(stream, dx, dy[, callback])` | `(...) => any` | `__readline.readline.moveCursor` | 📋 Planned | - |
| `readlinePromises.Interface` | `(...) => any` | `__readline.readlinePromises.Interface` | 📋 Planned | - |
| `readlinePromises.Readline` | `(...) => any` | `__readline.readlinePromises.Readline` | 📋 Planned | - |
| `readlinePromises.createInterface(options)` | `(...) => any` | `__readline.readlinePromises.createInterface` | 📋 Planned | - |
| `rl.clearLine(dir)` | `(...) => any` | `__readline.rl.clearLine` | 📋 Planned | - |
| `rl.clearScreenDown()` | `(...) => any` | `__readline.rl.clearScreenDown` | 📋 Planned | - |
| `rl.close()` | `(...) => any` | `__readline.rl.close` | 📋 Planned | - |
| `rl.commit()` | `(...) => any` | `__readline.rl.commit` | 📋 Planned | - |
| `rl.cursorTo(x[, y])` | `(...) => any` | `__readline.rl.cursorTo` | 📋 Planned | - |
| `rl.getCursorPos()` | `(...) => any` | `__readline.rl.getCursorPos` | 📋 Planned | - |
| `rl.getPrompt()` | `(...) => any` | `__readline.rl.getPrompt` | 📋 Planned | - |
| `rl.moveCursor(dx, dy)` | `(...) => any` | `__readline.rl.moveCursor` | 📋 Planned | - |
| `rl.pause()` | `(...) => any` | `__readline.rl.pause` | 📋 Planned | - |
| `rl.prompt([preserveCursor])` | `(...) => any` | `__readline.rl.prompt` | 📋 Planned | - |
| `rl.question(query[, options])` | `(...) => any` | `__readline.rl.question` | 📋 Planned | - |
| `rl.question(query[, options], callback)` | `(...) => any` | `__readline.rl.question` | 📋 Planned | - |
| `rl.resume()` | `(...) => any` | `__readline.rl.resume` | 📋 Planned | - |
| `rl.rollback()` | `(...) => any` | `__readline.rl.rollback` | 📋 Planned | - |
| `rl.setPrompt(prompt)` | `(...) => any` | `__readline.rl.setPrompt` | 📋 Planned | - |
| `rl.write(data[, key])` | `(...) => any` | `__readline.rl.write` | 📋 Planned | - |
| `rl[Symbol.asyncIterator]()` | `(...) => any` | `__readline.rl[Symbol.asyncIterator]` | 📋 Planned | - |
| `rl[Symbol.dispose]()` | `(...) => any` | `__readline.rl[Symbol.dispose]` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `readline` are organized per API under `internal/compiler/testdata/corpus/readline/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/readline/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
