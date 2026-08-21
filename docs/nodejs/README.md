# Node.js & ECMAScript Standard Library Checklist Specification

This document defines the architectural classification, official upstream sources of truth, standardized markdown checklist format, parity criteria, and automated generation workflow for all built-in symbols and standard library APIs documented under `docs/nodejs/`.

---

## 1. Directory & Category Architecture

All standard library symbols in `scriptgo` are strictly organized into four subdirectories under `docs/nodejs/`, corresponding directly to the `BuiltinCategory` enumeration defined in `internal/lowering/builtins.go`:

```text
docs/nodejs/
├── README.md               # Master specification & checklist authoring guide
├── ecmascript/             # CategoryECMAScript ("ECMAScript")
│   └── ...                 # ECMA-262 standards (Math, JSON, RegExp, Symbol, Date, Collections, etc.)
├── webcompat/              # CategoryWebCompat ("WebCompat")
│   └── ...                 # WinterCG / Server Web APIs (fetch, URL, Streams, Encoding, etc.)
├── nodeglobal/             # CategoryNodeGlobal ("NodeGlobal")
│   └── ...                 # Node.js global scope (console, process, Buffer, timers, etc.)
└── nodemodule/             # CategoryNodeModule ("NodeModule")
    └── ...                 # Node.js core modules imported via 'node:*' (fs, path, os, crypto, etc.)
```

---

## 2. The 2-Source Architecture Principle

All checklists in `docs/nodejs/` are strictly generated and synchronized using only **2 sources of truth**:

1. **API Catalog Source (Official Specifications)**:
   - **ECMAScript & Ambient Standards**: Dynamically parsed from official TypeScript `.d.ts` definitions bundled in `microsoft/typescript-go` (`lib.es5.d.ts`, `lib.es2015.*.d.ts`, `lib.es2020.*.d.ts`, `lib.es2024.*.d.ts`).
   - **Node.js Modules & Web Compat APIs**: Dynamically extracted from official upstream Node.js 22 LTS API JSON documentation (`https://nodejs.org/docs/latest-v22.x/api/<module>.json`).
   - **Zero Hardcoding**: Symbols, signatures, and properties are never manually written or hardcoded in generator Go structs.

2. **Implementation Status Source (Local Corpus Discovery)**:
   - Status is computed dynamically by discovering test cases under `internal/compiler/testdata/corpus/<feature>/<api_name>/`.
   - **`✅ Done`**: At least one test case (`main.ts`) exists under `internal/compiler/testdata/corpus/<feature>/<api_name>/` (e.g. `basic/`, `edge_cases/`).
   - **`📋 Planned`**: No test case exists under `internal/compiler/testdata/corpus/<feature>/<api_name>/` yet (field set to `-`).

---

## 2. Authoritative Sources of Truth & Parity Gates

To ensure complete, verifiable parity with **Node.js 22 LTS**, every category draws its type declarations, signatures, and semantic specifications from official upstream repositories:

| Category | Type Definition Source (TypeScript AST) | Specification Reference | Parity Gate / Test Oracle |
| :--- | :--- | :--- | :--- |
| **`ecmascript/`** | [microsoft/TypeScript `src/lib/lib.es*.d.ts`](https://github.com/microsoft/TypeScript/tree/main/src/lib) (`lib.es5`, `lib.es2015.*`, `lib.es2020.*`, `lib.es2024`) | [TC39 ECMA-262 Specification](https://tc39.es/ecma262/) | [TC39 Test262 Test Suite](https://github.com/tc39/test262) & TypeScript baselines |
| **`webcompat/`** | [microsoft/TypeScript `src/lib/lib.dom.d.ts`](https://github.com/microsoft/TypeScript/blob/main/src/lib/lib.dom.d.ts) (Server subset) | [WinterCG Minimum Common Web API](https://wintercg.org/) & [WHATWG Standards](https://spec.whatwg.org/) | [Web Platform Tests (WPT)](https://github.com/web-platform-tests/wpt) & Node.js WPT runner |
| **`nodeglobal/`** | [DefinitelyTyped `@types/node/globals.d.ts`](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node) (`process.d.ts`, `console.d.ts`, `buffer.d.ts`, `timers.d.ts`) | [Node.js 22 LTS Globals Docs](https://nodejs.org/docs/latest-v22.x/api/globals.html) | Node.js 22 LTS test suite (`test/parallel/test-global-*.js`) |
| **`nodemodule/`** | [DefinitelyTyped `@types/node/{fs,path,os,crypto,...}.d.ts`](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node) | [Node.js 22 LTS API Documentation](https://nodejs.org/docs/latest-v22.x/api/) | Node.js 22 LTS module test suite (`test/parallel/test-*.js`) |

### Scope & Import Boundary Summary

```text
┌──────────────────────────────────────────────────────────────┐
│                TypeScript / JavaScript Source                │
└──────────────────────────────┬───────────────────────────────┘
                               │
              ┌────────────────┴────────────────┐
              │                                 │
              ▼                                 ▼
┌──────────────────────────┐       ┌──────────────────────────┐
│   Language Environment   │       │     Node Environment     │
└────────────┬─────────────┘       └────────────┬─────────────┘
             │                                  │
             ▼                    ┌─────────────┼─────────────┐
┌──────────────────────────┐      ▼             ▼             ▼
│ 1. ECMAScript Built-ins  │ ┌──────────┐ ┌──────────┐ ┌─────────────┐
│                          │ │ 2. Web   │ │ 3. Node  │ │ 4. Built-in │
│ Global: NO import        │ │ globals  │ │ globals  │ │ modules     │
│                          │ │          │ │          │ │             │
│ Array, Object, Math      │ │ Global:  │ │ Global:  │ │ Import: YES │
│ JSON, Promise, Map       │ │ NO import│ │ NO import│ │ node:fs     │
│ Set, RegExp, Date, ...   │ │          │ │          │ │ node:path   │
└──────────────────────────┘ │ fetch    │ │ process  │ │ node:os     │
                             │ URL      │ │ Buffer   │ │ node:crypto │
                             │ Streams  │ │ console  │ │ node:http   │
                             └──────────┘ └──────────┘ └─────────────┘
```

---

## 3. Standard Checklist Markdown Format

Every checklist file created in any subfolder (`ecmascript/`, `webcompat/`, `nodeglobal/`, `nodemodule/`) must adhere to this uniform 5-section template.

### 3.1. Markdown Template

```markdown
# [Feature / Module Name] Implementation Checklist

> **Category**: `CategoryECMAScript` | `CategoryWebCompat` | `CategoryNodeGlobal` | `CategoryNodeModule`  
> **Import Path**: `N/A (Global Scope)` | `node:<module_name>` (e.g. `node:fs`, `node:path`)  
> **Specification Reference**: [Link to official ECMA-262 / WHATWG / Node.js 22 LTS documentation]  
> **Type Definition Source**: [Link to official lib.es*.d.ts or @types/node definition file]  
> **Gate Oracle**: [Test262 / WPT / Node.js 22 LTS Test Suite]

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Whether symbols are global or require module import.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `exampleMethod(a, b)` | `(a: string, b: number) => boolean` | `__module.exampleMethod` | ✅ Done | `internal/compiler/testdata/corpus/module/basic/` |
| `pendingMethod()` | `() => void` | `__module.pendingMethod` | ⏳ In Progress | `internal/compiler/testdata/corpus/module/pending/` |
| `futureMethod()` | `() => number` | `__module.futureMethod` | 📋 Planned | - |
| `unsupportedDynamic()` | `(...args: any[]) => any` | - | 🚫 Excluded | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
If the API is accessible both globally and through a module import (e.g., `Buffer` vs `node:buffer`, `process` vs `node:process`), verify that both surfaces resolve to the identical lowering implementation.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Regression Test**: Add a regression test under `internal/compiler/testdata/corpus/` with `main.ts` and `run.expected` (or `run.err` / `check.err`).
- [ ] **Step 7: Documentation Sync**: Update the status to `✅ Done` in this checklist and synchronize metrics in `docs/typescript-parity-report.md`.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
```

---

## 4. Status Badges & Criteria

To maintain clarity and auditability across all files, use exclusively the following status badges:

| Status Badge | Name | Criteria |
| :---: | :--- | :--- |
| ✅ **Done** | Complete | Full pipeline implemented (Frontend → Lowering → IR → Interpreter → LLVM/Clang) with passing regression tests in `internal/compiler/testdata/corpus/`. |
| ⏳ **In Progress** | In Progress | Active implementation; lowered or executed in interpreter, but pending complete LLVM native runtime integration or full test coverage. |
| 📋 **Planned** | Planned | Symbol identified, typed in standard definitions, and scheduled in roadmap, but lowering/runtime code has not yet been written. |
| 🚫 **Excluded** | Out of Scope | Explicitly excluded from the native AOT subset (e.g., browser DOM manipulation, dynamic `eval()`, runtime `Function()` generation). |

---

## 5. Repository Maintenance Invariants

1. **Acyclic Module Boundaries**:
   Documentation must strictly respect the module responsibilities defined in `AGENTS.md` and `docs/application-structure.md`. Frontend type checking, lowering logic, IR models, and native backends must remain decoupled.
2. **Dual-Surface Consistency**:
   APIs that exist across multiple surfaces (e.g., `crypto.randomUUID()` in WebCompat/NodeGlobal vs `import { randomUUID } from "node:crypto"`) must resolve to the identical lowering target and C ABI callee.
3. **Atomic Parity Sync**:
   Every PR or commit that introduces a new built-in symbol or changes parity status must atomically update:
   - The specific checklist file under `docs/nodejs/<category>/`.
   - The central parity report at `docs/typescript-parity-report.md`.
