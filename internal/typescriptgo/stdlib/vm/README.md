# Node.js Virtual Machine (`node:vm`) Module

This directory contains the standard library implementation and type definitions for Node.js `node:vm` in ScriptGo.

## Architecture & Tier Boundaries

ScriptGo uses explicit compilation tiers as specified in [`docs/compilation-tiers.md`](../../../../docs/compilation-tiers.md).

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                            ScriptGo `node:vm`                               │
├──────────────────────────────────────┬──────────────────────────────────────┤
│             Static Tier              │             Dynamic Tier             │
│              (Default)               │             (`--dynamic`)            │
├──────────────────────────────────────┼──────────────────────────────────────┤
│ - Complete TypeScript types & options│ - Full runtime JavaScript evaluation │
│ - Static AST / type-checking support │ - Execution in QuickJS-ng dynamic    │
│ - Safe compile-time contract shims   │   islands / sandbox contexts         │
│ - Zero runtime engine bloat in AOT   │ - Dynamic `eval`, `compileFunction`, │
│   machine binaries                   │   and `Script.runInContext` support  │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

### 1. Static Tier (Default AOT Compilation)
- **Role**: Provides complete TypeScript definitions, interfaces, classes (`Script`, `Module`, `SourceTextModule`, `SyntheticModule`), options (`ScriptOptions`, `CreateContextOptions`, `RunningScriptOptions`, etc.), and constants (`DONT_CONTEXTIFY`, `constants`).
- **Design Rationale**: A pure AOT native binary produced by ScriptGo compiles statically to LLVM IR without embedding a JavaScript interpreter. Arbitrary dynamic string execution at runtime cannot be evaluated directly by native machine instructions. Static shims ensure that TypeScript code and libraries importing `node:vm` compile and type-check cleanly without frontend diagnostics.

### 2. Dynamic Tier (`--dynamic`)
- **Role**: When building or running with `--dynamic`, dynamic code execution calls are directed to **QuickJS-ng dynamic islands**.
- **Execution Model**: QuickJS-ng provides isolated runtime contexts and ES module records to parse, compile, link, and evaluate dynamic JavaScript code strings at runtime.

---

## Supported APIs & Symbols

### Classes
- `Script`: Represents precompiled scripts (`runInContext`, `runInNewContext`, `runInThisContext`, `createCachedData`).
- `Module`: Base class for ECMAScript module records (`status`, `identifier`, `namespace`, `error`, `link`, `evaluate`).
- `SourceTextModule`: ECMAScript source text module (`dependencySpecifiers`, `moduleRequests`, `linkRequests`, `instantiate`).
- `SyntheticModule`: Synthetic module with programmable exports (`setExport`).

### Functions
- `compileFunction(code, params?, options?)`: Compiles code into a callable function.
- `createContext(contextObject?, options?)`: Contextifies an object.
- `isContext(object)`: Checks if an object is contextified.
- `measureMemory(options?)`: Measures VM memory usage.
- `runInContext(code, contextifiedObject, options?)`: Runs code in a context.
- `runInNewContext(code, contextObject?, options?)`: Creates context and runs code.
- `runInThisContext(code, options?)`: Runs code in the current global context.

### Constants
- `DONT_CONTEXTIFY` (1)
- `constants.USE_MAIN_CONTEXT_DEFAULT_LOADER` (0)
- `constants.DONT_CONTEXTIFY` (1)
