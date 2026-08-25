# Standard Library Compatibility & Parity Architecture

`scriptgo` targets a standard library surface compatible with modern JavaScript
runtimes (ECMAScript standards, Node.js 22 LTS, and WinterCG server APIs) so
useful backend, CLI, and systems TypeScript code can compile to efficient
native binaries.

TypeScript-Go remains responsible for parsing, binding, module resolution, and
type checking; `scriptgo` owns native eligibility, runtime mapping, and parity
verification.

---

## The Four Standard Library Tiers

To maintain clear module boundaries and enable robust **Gate Parity Checking**,
all built-ins and runtime symbols are categorized into four distinct groups:

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
│ Array, Object, Math      │ │ globals  │ │ globals  │ │ modules     │
│ JSON, Promise, Map       │ │          │ │          │ │             │
│ Set, RegExp, Date, ...   │ │ fetch    │ │ process  │ │ node:fs     │
└──────────────────────────┘ │ URL      │ │ Buffer   │ │ node:path   │
                             │ Request  │ │ console* │ │ node:os     │
                             │ Response │ │ setImm...│ │ node:crypto │
                             │ Abort... │ │ ...      │ │ node:http   │
                             └──────────┘ └──────────┘ └─────────────┘
```

| Category | Examples | Scope & Meaning | Import Required? | Parity Test Oracle (Gate) |
| :--- | :--- | :--- | :--- | :--- |
| **1. ECMAScript built-ins** | `Array`, `Object`, `Promise`, `Math`, `JSON`, `Map`, `Set`, `Error`, `Number`, `String`, `NaN`, `Infinity` | Defined strictly by ECMA-262 specifications. Not tied to Node.js or browser hosts. | **NO (Global Scope)** | **ECMA Test262** test suite & TypeScript baselines |
| **2. Web-compatible globals** | `fetch`, `Headers`, `Request`, `Response`, `URL`, `URLSearchParams`, `AbortController`, `AbortSignal`, `TextEncoder`, `TextDecoder`, `Blob`, `structuredClone`, `setTimeout`, `queueMicrotask`, `performance` | Standard Web APIs adopted globally across server runtimes (WinterCG / Node.js). **Excludes browser DOM APIs.** | **NO (Global Scope)** | **WPT (Web Platform Tests)** & Node.js Web API suite |
| **3. Node-specific globals** | `process`, `Buffer`, `setImmediate`, `clearImmediate`, `__dirname`, `__filename`, `global`, `console` | Host environment APIs injected by Node.js runtime bootstrap. | **NO (Global Scope)** | **Node.js 22 LTS** globals test suite |
| **4. Node built-in modules** | `node:fs`, `node:path`, `node:crypto`, `node:os`, `node:events`, `node:util`, `node:process`, `node:child_process` | Standard library modules explicitly imported via `node:*` or bare specifiers. | **YES (`import` / `require`)** | **Node.js 22 LTS** `test/parallel/test-*.js` suites |

---

## Scope & Import Boundaries

### 1. Global Surfaces (Categories 1, 2, 3) — Zero Import Required
Any symbol in Categories 1, 2, or 3 is available in every file automatically without an `import` statement:
- **ECMAScript:** `Math.max(1, 2)`, `JSON.parse(str)`, `const arr = new Array()`
- **Web Globals:** `const url = new URL(input)`, `await fetch(url)`, `setTimeout(fn, 100)`
- **Node Globals:** `console.log("log")`, `process.cwd()`, `Buffer.from("abc")`

### 2. Module Surfaces (Category 4) — Explicit Import Required
Symbols in Category 4 are encapsulated within standard modules and **cannot** be called from the global scope. They require an explicit `import` statement:
```ts
// Canonical Node.js import (Recommended)
import * as fs from "node:fs";
import * as crypto from "node:crypto";
import { join, dirname } from "node:path";

// Bare specifier compatibility alias
import * as path from "path";
```

### 3. Dual-Surface Resolution Rules (Global vs Module)

Certain concepts exist in both the global scope and as a dedicated module. `scriptgo` enforces the following semantic boundary:

| Symbol / API | Global Usage (No Import) | Module Usage (`import ... from 'node:*'`) |
| :--- | :--- | :--- |
| **`process`** | Category 3 Global: `process.argv`, `process.cwd()`, `process.exit()` | Category 4 Module: `import process from "node:process"` (ESM specifier alias to global instance) |
| **`crypto`** | *Not available globally* | Category 4 Module: `import * as crypto from "node:crypto"` (`crypto.randomUUID()`, etc.) |
| **`Buffer` / `buffer`** | Category 3 Global: `Buffer.from()`, `Buffer.alloc()` | Category 4 Module: `import { Buffer } from "node:buffer"` (Module-level exports, Blob/constants) |
| **Timers / `timers`** | Category 2 Web Global: `setTimeout`, `setInterval`, `clearTimeout` | Category 4 Module: `import { setTimeout } from "node:timers/promises"` (Promise-based timer variants) |

---


## Web-Compatible Globals: Server Focus & Frontend Exclusion

`scriptgo` is a native backend/systems compiler targeting servers, CLIs, and
high-performance tools. It is **not** a browser engine and does **not** support
DOM layout or UI rendering.

### ✅ IN-SCOPE: Server & Systems Web Globals (WinterCG / Node.js Compatible)

1. **HTTP & Networking:** `fetch`, `Headers`, `Request`, `Response`, `FormData`
2. **URL & Routing:** `URL`, `URLSearchParams`, `URLPattern`
3. **I/O Streams:** `ReadableStream`, `WritableStream`, `TransformStream`, `ByteLengthQueuingStrategy`
4. **Cancellation & Signals:** `AbortController`, `AbortSignal`
5. **Binary & Text Encoding:** `TextEncoder`, `TextDecoder`, `Blob`, `File`
6. **Base64 & Utilities:** `btoa`, `atob`, `structuredClone`
7. **Events Core:** `Event`, `EventTarget`, `CustomEvent`
8. **Timers & Microtasks:** `setTimeout`, `clearTimeout`, `setInterval`, `clearInterval`, `queueMicrotask`
9. **High-Resolution Time:** `performance` (`performance.now()`, `PerformanceMark`)
10. **Real-time Comms (Future):** `WebSocket` client

### ❌ OUT-OF-SCOPE: Frontend & Browser DOM APIs (Explicitly Excluded)

These APIs will **not** be admitted to the native runtime:
- **DOM Tree & HTML Elements:** `window`, `document`, `HTMLElement`, `Element`, `Node`, `HTMLDivElement`, `ShadowRoot`, `DocumentFragment`.
- **Browser Client Storage & History:** `localStorage`, `sessionStorage`, `indexedDB`, `history`, `location`.
- **Graphics & CSS Rendering:** `requestAnimationFrame`, `Canvas`, `WebGL`, `CSSStyleDeclaration`, `IntersectionObserver`, `MutationObserver`, `ResizeObserver`.
- **Browser Hardware / UI Events:** `MouseEvent`, `KeyboardEvent`, `TouchEvent`, `AudioContext`, `Geolocation`, `Notification`, `ServiceWorker`, `WebRTC` (`RTCPeerConnection`).

---

## Three-Tier Compilation Policy

The standard library participates in the same three-tier compilation policy as
application code (see [`compilation-tiers.md`](compilation-tiers.md)):

- **Static:** Pure TypeScript stdlib or a versioned native primitive whose
  behavior is proven and compiled directly into the native executable.
- **Dynamic:** JavaScript/npm library code that requires full JavaScript dynamic
  semantics or reflection, executed through embedded QuickJS-ng when `--dynamic` is enabled.
- **Unsupported:** An API with no implemented Static or Dynamic contract;
  compilation fails with a source-anchored diagnostic.

---

## Gate Parity Checking Rules

Every standard-library surface admitted to the compiler must satisfy the
following parity verification criteria:

1. **Documented Reference:** Pin parity claims to ECMAScript 2022+ and Node.js 22 LTS.
2. **Oracle Verification:**
   - ECMA Built-ins: Must pass matching Test262 test cases.
   - Web Globals: Must pass relevant WPT tests without browser-specific DOM assumptions.
   - Node Globals & Modules: Must match Node.js 22 LTS observable stdout, exit codes, exceptions, and error messages.
3. **Multi-Target Coverage:**
   - Reference interpreter test fixture (`internal/interpreter`).
   - Native LLVM compiled executable test fixture (`internal/compiler/testdata/corpus/stdlib/`).
4. **Differential Parity Gate:** CI checks run differential assertions verifying
   identical outputs between `node <test>.ts` and `scriptgo run <test>.ts`.

---

## Module Names & Specifier Resolution

Both bare specifiers and canonical `node:` prefixed specifiers are supported:

```ts
import * as path from "node:path"; // Canonical Node.js style
import * as fs from "fs";          // Bare compatibility alias
```

Lowering and the module loader canonicalize module specifiers while preserving
accurate source spans in compiler diagnostics.

---

## Stdlib Implementation Hierarchy

Implement standard-library functionality following this hierarchy:

```text
Native C ABI Primitives -> Core TypeScript stdlib -> High-level TypeScript stdlib
```

- **Pure TypeScript Single Source of Truth:** All standard library modules (`fs`, `net`, `http`, `stream`, `crypto`, `os`, `path`, `events`, `url`, `buffer`, `timers`, `process`, `child_process`) are authored directly as `.ts` files with full TypeScript type annotations, interfaces, classes, and exported functions.
- **Zero `.d.ts` Dual Maintenance:** TypeScript-Go infers all types, signatures, and contracts directly from the `.ts` module sources loaded into the virtual module filesystem (`node_modules/<name>/index.ts`). Redundant module `.d.ts` files and code generation scripts are completely eliminated.
- **Thin TS Wrappers + Minimal C Runtime:** High-level APIs (overload handling, option validation, stream piping, event dispatching, URL formatting) are written 100% in pure TypeScript, invoking low-level C runtime intrinsics (`internal/runtime/`) only for POSIX system calls (sockets, file descriptors, crypto RNG, thread pools).

