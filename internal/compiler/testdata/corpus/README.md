# ScriptGo Compiler Test Corpus

This directory contains the test corpus used for testing the ScriptGo compiler, lowering, interpreter, and native LLVM code generation.

---

## 1. Directory Structure

```text
internal/compiler/testdata/corpus/
├── api/          # 1-API test cases strictly mapped to docs/nodejs/ checklists
├── scenarios/    # Multi-API integration workflows, pipelines, and composite tests
└── language/     # Core TypeScript/JavaScript syntax and language constructs
```

---

## 2. Rule for `corpus/api/` Folder Naming

> [!IMPORTANT]
> Every directory under `internal/compiler/testdata/corpus/api/<feature>/<api_name>/` **MUST strictly match the exact API symbol name** as documented in `docs/nodejs/`.

### Directory Hierarchy:
```text
internal/compiler/testdata/corpus/api/<feature>/<api_name>/<test_case>/main.ts
```

### Examples:
| Feature | API Symbol in `docs/nodejs` | Corpus Directory Path |
| :--- | :--- | :--- |
| `array` | `Array.prototype.indexOf` | `internal/compiler/testdata/corpus/api/array/indexOf/basic/main.ts` |
| `array` | `Array.prototype.forEach` | `internal/compiler/testdata/corpus/api/array/forEach/basic/main.ts` |
| `fs` | `fs.readFileSync` | `internal/compiler/testdata/corpus/api/fs/readFileSync/basic/main.ts` |
| `fs` | `fs.writeFileSync` | `internal/compiler/testdata/corpus/api/fs/writeFileSync/basic/main.ts` |
| `crypto` | `crypto.randomUUID` | `internal/compiler/testdata/corpus/api/crypto/randomUUID/basic/main.ts` |
| `buffer` | `Buffer.allocUnsafe` | `internal/compiler/testdata/corpus/api/buffer/allocUnsafe/basic/main.ts` |
| `object` | `Object.hasOwn` | `internal/compiler/testdata/corpus/api/object/hasOwn/basic/main.ts` |
| `number` | `Number.isInteger` | `internal/compiler/testdata/corpus/api/number/isInteger/basic/main.ts` |
| `timers` | `timers.setImmediate` | `internal/compiler/testdata/corpus/api/timers/setImmediate/basic/main.ts` |
| `events` | `events.EventEmitter` | `internal/compiler/testdata/corpus/api/events/EventEmitter/basic/main.ts` |

### Multi-Test Support:
Each API directory can contain multiple subdirectories for different test scenarios:
```text
internal/compiler/testdata/corpus/api/array/map/
├── basic/
│   ├── main.ts
│   └── run.expected
└── edge_cases/
    ├── main.ts
    └── run.expected
```

---

## 3. Automated Parity Check

The documentation generator `scripts/gendocs/main.go` scans `internal/compiler/testdata/corpus/api/<feature>/<api_name>/`:
- If at least one `main.ts` test case exists: marked **`✅ Done`** with the test path linked in the markdown checklist.
- If no test case exists: marked **`📋 Planned`** with `-`.
