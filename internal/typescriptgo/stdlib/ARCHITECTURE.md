# ScriptGo Standard Library Architecture & Layout Specification

> **Location**: `internal/typescriptgo/stdlib/`  
> **Upstream Reference Repositories**:
>
> - **Node.js v22 Core Implementation**: [`nodejs/node:lib/` (v22.x)](https://github.com/nodejs/node/tree/v22.x/lib)
> - **DefinitelyTyped Node.js v22 Contracts**: [`DefinitelyTyped/types/node/v22/`](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node/v22)
> - **ScriptGo Architecture Rules**: [`AGENTS.md`](../../../AGENTS.md) & [`docs/application-structure.md`](../../../docs/application-structure.md)

---

## 1. Upstream Empirical Investigation (Fetched Live)

By directly inspecting the GitHub API for [`nodejs/node:lib` (v22.x)](https://github.com/nodejs/node/tree/v22.x/lib) and [`DefinitelyTyped:types/node/v22`](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node/v22), the official module topology of Node.js v22 is structured as follows:

### 1.1 Node.js Core (`nodejs/node:lib/`) Structure

In Node.js v22:

- **Top-level module entry points** are paired with **subpath directories**:
  - `assert.js` + `assert/strict.js`
  - `dns.js` + `dns/promises.js`
  - `fs.js` + `fs/promises.js`
  - `inspector.js` + `inspector/promises.js`
  - `path.js` + `path/posix.js` + `path/win32.js`
  - `readline.js` + `readline/promises.js`
  - `stream.js` + `stream/consumers.js` + `stream/promises.js` + `stream/web.js`
  - `test.js` + `test/reporters.js`
  - `timers.js` + `timers/promises.js`
  - `util.js` + `util/types.js`
- **Internal implementation files** (`lib/internal/*`) provide private JS helpers in Node.js. In ScriptGo, native operations bypass interpreted JS helpers and are handled directly in Go lowering (`internal/lowering/`) and C runtime ABI (`internal/runtime/native/`).

### 1.2 DefinitelyTyped (`types/node/v22/`) Structure

In DefinitelyTyped:

- Every root module has a `.d.ts` declaration file (e.g. `fs.d.ts`, `stream.d.ts`, `assert.d.ts`).
- Dedicated subpath folders mirror Node.js submodules:
  - `assert/strict.d.ts`
  - `dns/promises.d.ts`
  - `fs/promises.d.ts`
  - `readline/promises.d.ts`
  - `stream/consumers.d.ts`, `stream/promises.d.ts`, `stream/web.d.ts`
  - `timers/promises.d.ts`
- Ambient modules for both prefixes (`node:<module>` and `<module>`) are declared in `index.d.ts`.

---

## 2. Full Target Layout for ScriptGo Standard Library

Designed to achieve 100% architectural symmetry with Node.js v22 and DefinitelyTyped, without empty directories or unused placeholders, while adhering to ScriptGo's file length guideline (~500–700 lines max):

```text
internal/typescriptgo/stdlib/
├── ARCHITECTURE.md                 # Architecture, standards mapping & layout documentation
├── globals.d.ts                    # Global standard library definitions (fetch, Response, console)
│
├── assert/                         # node:assert subpath family
│   ├── index.ts                    # node:assert (assert function, AssertionError, legacy aliases)
│   └── strict.ts                   # node:assert/strict (strict deep equality mode)
│
├── dns/                            # node:dns subpath family
│   ├── index.ts                    # node:dns (callback APIs: lookup, resolve4, resolve6, reverse)
│   └── promises.ts                 # node:dns/promises (Promise-based DNS resolution)
│
├── fs/                             # node:fs subpath family
│   ├── index.ts                    # node:fs (Sync & Callback APIs: readFileSync, Stats, Dirent, Dir)
│   └── promises.ts                 # node:fs/promises (FileHandle, Promise-based file system APIs)
│
├── inspector/                      # node:inspector subpath family
│   ├── index.d.ts                  # node:inspector (V8 inspector protocol session)
│   └── promises.d.ts               # node:inspector/promises (Promise-based inspector session)
│
├── path/                           # node:path subpath family
│   ├── index.ts                    # node:path (auto platform detection: sep, join, resolve)
│   ├── posix.ts                    # node:path/posix (strict POSIX separator /)
│   └── win32.ts                    # node:path/win32 (strict Windows separator \)
│
├── readline/                       # node:readline subpath family
│   ├── index.ts                    # node:readline (Interface, createInterface, emitKeypressEvents)
│   └── promises.ts                 # node:readline/promises (Promise-based async readline interface)
│
├── stream/                         # node:stream subpath family (Implemented)
│   ├── index.ts                    # node:stream (Stream, Readable, Writable, Transform, Duplex)
│   ├── consumers.ts                # node:stream/consumers (buffer, text, json, arrayBuffer, blob)
│   ├── promises.ts                 # node:stream/promises (pipeline, finished)
│   └── web.ts                      # node:stream/web & webstreams (WHATWG Streams standard)
│
├── test/                           # node:test subpath family (Node.js Test Runner)
│   ├── index.ts                    # node:test (describe, it, test, suite, before, after)
│   └── reporters.ts                # node:test/reporters (spec, tap, dot, junit)
│
├── timers/                         # node:timers subpath family
│   ├── index.ts                    # node:timers (setTimeout, setInterval, setImmediate, clear*)
│   └── promises.ts                 # node:timers/promises (setTimeout, setInterval async iterator)
│
├── util/                           # node:util subpath family
│   ├── index.ts                    # node:util (promisify, format, inspect, deprecate, MIME)
│   └── types.ts                    # node:util/types (type-testing predicates: isAnyArrayBuffer, etc.)
│
├── vm/                             # node:vm subpath family (Implemented)
│   ├── README.md                   # Multi-tier compilation boundary specification
│   └── index.d.ts                  # node:vm ambient declarations (Script, Module, createContext)
│
├── wasi/                           # node:wasi subpath family
│   └── index.ts                    # node:wasi (WASI system integration & WASM host APIs)
│
├── worker_threads/                 # node:worker_threads subpath family
│   └── index.ts                    # node:worker_threads (Worker, parentPort, MessageChannel)
│
├── async_hooks.ts                  # node:async_hooks
├── atomics.ts                      # Atomics & SharedArrayBuffer shims
├── buffer.ts                       # node:buffer (Buffer, Blob, File, SlowBuffer)
├── child_process.ts                # node:child_process (spawn, exec, fork)
├── cluster.ts                      # node:cluster (multi-process worker management)
├── console.ts                      # node:console
├── constants.ts                    # node:constants (system, signal, and crypto constants)
├── crypto.ts                       # node:crypto (Hmac, Hash, Cipher, WebCrypto shims)
├── dgram.ts                        # node:dgram (Socket, createSocket)
├── diagnostics_channel.ts          # node:diagnostics_channel (Channel, tracing)
├── domain.ts                       # node:domain (Domain event dispatch)
├── events.ts                       # node:events (EventEmitter, EventTarget, once, on)
├── http.ts                         # node:http (Server, ClientRequest, IncomingMessage, Agent)
├── http2.ts                        # node:http2 (HTTP/2 sessions and framing)
├── https.ts                        # node:https (HTTPS server and client options)
├── module.ts                       # node:module (createRequire, builtinModules, SourceMap)
├── net.ts                          # node:net (Socket, Server, IPC)
├── os.ts                           # node:os (cpus, arch, platform, homedir)
├── perf_hooks.ts                   # node:perf_hooks (Performance, PerformanceObserver)
├── process.ts                      # node:process (argv, env, exit, memoryUsage)
├── punycode.ts                     # node:punycode (Punycode conversion)
├── querystring.ts                  # node:querystring (parse, stringify)
├── sqlite.ts                       # node:sqlite (DatabaseSync, Session)
├── string_decoder.ts               # node:string_decoder (StringDecoder)
├── tls.ts                          # node:tls (TLSSocket, Server, SecureContext)
├── trace_events.ts                 # node:trace_events (V8 tracing)
├── tty.ts                          # node:tty (ReadStream, WriteStream)
├── url.ts                          # node:url (URL, URLSearchParams, fileURLToPath)
├── v8.ts                           # node:v8 (heap statistics and serialization)
├── weak.ts                         # WeakRef & FinalizationRegistry shims
├── webcrypto.ts                    # Web Cryptography API
└── zlib.ts                         # node:zlib (Gzip, Deflate, Brotli, Inflate)
```

---

## 3. Strict Rules & Architectural Contracts

### 3.1 Type Safety & Static Mode Compliance

- **No `any`**: The `any` type triggers compile error `SG1001` in static mode. Every API must specify concrete types, generic parameters, or union types.
- **Controlled `unknown`**: `unknown` is only acceptable where values are genuinely untyped prior to dynamic inspection (e.g. `JSON.parse` output, `event` payload, or `catch (err: unknown)`).
- **Ambient & Subpath Resolution**:
  - `internal/typescriptgo/config.go` (`buildVirtualEnvironment`) exposes all subpath modules to both `"node:<module>"` and `"<module>"` import forms.
  - Subpaths with `/index.ts` automatically resolve when imported as `"node:<folder>"` (e.g. `import * as fs from "node:fs"` resolves to `stdlib/fs/index.ts` or `stdlib/fs.ts`).

### 3.2 Maintenance Policy

- **Parity verification**: Every addition or modification must be cross-verified using `go run ./cmd/parity` against Node.js v22, maintaining the 1:1 behavioral oracle.
