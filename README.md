<div align="center">
  <h1>ScriptGo</h1>
</div>

<div align="center">

[![CI](https://github.com/pilotworks/scriptgo/actions/workflows/ci.yml/badge.svg)](https://github.com/pilotworks/scriptgo/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org)
[![TypeScript Parity](<https://img.shields.io/badge/TypeScript%20Parity-96.6%25%20(366%2F379)-success.svg>)](docs/typescript-parity-report.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platforms](<https://img.shields.io/badge/Platforms-macOS%20%7C%20Linux%20%7C%20Windows%20%7C%20WASM%20(WASI)-lightgrey.svg>)](#toolchain--cross-compilation)

**A high-performance Ahead-Of-Time (AOT) compiler compiling TypeScript to native standalone executables and WebAssembly (WASI) modules with Node.js parity.**

</div>

---

**ScriptGo** is a high-performance native compiler that runs TypeScript and JavaScript with Node.js-compatible semantics while compiling eligible code directly to standalone native binaries or WebAssembly modules. It combines the official [TypeScript (Go implementation)](https://github.com/microsoft/TypeScript) compiler frontend for parsing, type-checking, and diagnostics with an independent **Typed IR** system and an **LLVM IR / Native Machine Code** backend.

---

## Highlights & Features

- **High-Performance AOT Compilation**: Compiles TypeScript directly to native machine code (Mach-O, ELF, PE) and WebAssembly (`.wasm`) via LLVM.
- **Node.js Semantic Parity**: High parity across the 379-case regression test corpus checked against Node.js v22+ (366/379 native pass rate, 96.6%).
- **WebAssembly / WASI Target**: First-class Ahead-Of-Time compilation to standalone `.wasm` executables with `--target wasm32-wasi`, runnable across Node.js WASI, Wasmtime, Wasmer, and Edge runtimes.
- **Zero-Dependency Native Builds**: Automatically uses system `clang` or auto-detects `zig cc` for hassle-free out-of-the-box compilation and seamless cross-compilation (macOS, Linux, Windows, WASM).
- **Fast Execution**: Instantly compiles and runs scripts directly or produces optimized standalone binary builds.
- **Modern TypeScript & ECMAScript (ES2022 - ES2024)**:
  - **Explicit Resource Management**: Full `using` and `await using` resource disposal with `Symbol.dispose` and `Symbol.asyncDispose`.
  - **ES2024 Set Methods**: `union()`, `intersection()`, `difference()`, `symmetricDifference()`, `isSubsetOf()`, `isSupersetOf()`, `isDisjointFrom()`.
  - **ES2024 Utilities**: `Promise.withResolvers()`, `Object.groupBy()`, `Map.groupBy()`, `Array.fromAsync()`.
  - **Types & Primitives**: `number` (IEEE-754), `bigint`, `string` (UTF-8), `boolean`, `symbol` (with Symbol Registry), `null`, `undefined`, `unknown` (with type narrowing), Tuples, Enums (numeric, string, reverse mappings), Union types (`T | null | undefined`), Monomorphized Generics.
  - **Control Flow**: `if`/`else`, `switch`/`case` (with fallthrough), `while`, `do..while`, `for`, `for..of`, `for..in`, `for await..of`, Labeled statements (`break label`, `continue label`), `try`/`catch`/`finally`, `throw`, Array & Object destructuring, Spread/Rest (`...`), Tagged template literals, Optional chaining & calls (`?.`, `fn?.()`).
  - **Functions & Closures**: Lexical closures, arrow functions, default/optional/rest parameters, Generators (`function*`, `yield`, `yield*`), Async Generators.
  - **OOP & Classes**: Constructors, properties, static fields/methods, Class Static Blocks (`static { ... }`), Getters/Setters, Inheritance (`extends`, `super`), Polymorphic VTables, `instanceof`.
  - **Async Runtime**: `Promise` (resolve, reject, chaining), `async`/`await`, microtask queue execution conforming to JavaScript event loop ordering.
  - **Web Standards & WinterCG**: Streaming `fetch()` & WHATWG Streams (`ReadableStream`, `WritableStream`, `TransformStream`), `URL`, `URLSearchParams`, `TextEncoder`/`TextDecoder`, `AbortController`/`AbortSignal`.
  - **Node.js Standard Library**: High-performance native implementations for core Node.js modules (`node:fs`, `node:path`, `node:os`, `node:process`, `node:crypto`, `node:buffer`, `node:http`, `node:net`, `node:dgram`, `node:events`, `node:stream`, `node:assert`, `node:util`, `node:timers`, `node:zlib`, `node:tls`, `node:sqlite`). All placeholder/dummy stubs strictly removed.

---

## Documentation

- [`docs/typescript-parity-report.md`](docs/typescript-parity-report.md) - Comprehensive TypeScript/Node.js feature matrix and parity test report.
- [`docs/webassembly-wasi-architecture.md`](docs/webassembly-wasi-architecture.md) - WebAssembly & WASI compilation architecture and execution pipeline.
- [`docs/native-optimization-pipeline.md`](docs/native-optimization-pipeline.md) - Multi-level Dead Code Elimination (DCE), Link-Time Optimization (LTO), and Memory Layout.
- [`docs/native-subset.md`](docs/native-subset.md) - Native static subset definition and compatibility constraints.
- [`docs/compilation-tiers.md`](docs/compilation-tiers.md) - Static, Dynamic (QuickJS-ng island), and Unsupported compilation policy.
- [`docs/application-structure.md`](docs/application-structure.md) - Repository architecture, package ownership, and dependency direction.
- [`docs/typescript-to-native.md`](docs/typescript-to-native.md) - Pipeline boundaries, IR verification invariants, and LLVM lowering.
- [`docs/stdlib.md`](docs/stdlib.md) - Standard library API scope and Node.js runtime emulation policy.
- [`docs/roadmap.md`](docs/roadmap.md) - Project roadmap, milestones, and acceptance criteria.

---

## CLI Usage

### Running Code

```sh
# Compile and run immediately on host
scriptgo run examples/hello.ts

# Run with inline TypeScript code
scriptgo run -e "console.log('Hello from ScriptGo!')"
```

### Building Standalone Executables & WebAssembly Modules

```sh
# 1. Build a native binary for the host platform
scriptgo build examples/hello.ts -o hello

# 2. Build a WebAssembly (WASI) module
scriptgo build --target wasm32-wasi examples/hello.ts -o hello.wasm

# 3. Execute the generated WASM module via Node.js WASI or Wasmtime
node -e 'const { WASI } = require("wasi"); const fs = require("fs"); const wasi = new WASI({ version: "preview1", args: ["hello.wasm"], returnOnExit: true }); const bytes = fs.readFileSync("hello.wasm"); (async () => { const mod = await WebAssembly.compile(bytes); const inst = await WebAssembly.instantiate(mod, wasi.getImportObject()); wasi.start(inst); })();'
# Or via wasmtime
wasmtime hello.wasm

# 4. Build with debug symbols
scriptgo build examples/hello.ts --debug -o hello-debug

# 5. Build with Clang sanitizers (address, undefined, leak)
scriptgo build examples/hello.ts --sanitize address,undefined -o hello-sanitized
```

### Type Checking & Emitting IR

```sh
# Type-check and validate against the native subset
scriptgo check examples/hello.ts

# Emit LLVM IR
scriptgo emit examples/hello.ts -o hello.ll

# Emit verified Typed IR
scriptgo emit examples/hello.ts --mode typed-ir -o hello.ir
```

---

## Showcase Examples

Explore the [`examples/`](examples/) directory for complete TypeScript samples:

- [`examples/hello.ts`](examples/hello.ts) - Basic hello world script.
- [`examples/fibonacci.ts`](examples/fibonacci.ts) - Recursive and iterative performance benchmark.
- [`examples/classes_oop.ts`](examples/classes_oop.ts) - OOP with inheritance, class static blocks, getters/setters, and `instanceof`.
- [`examples/functional_arrays.ts`](examples/functional_arrays.ts) - Functional `Array` methods (`map`, `filter`, `reduce`, `find`, `some`, `every`).
- [`examples/advanced_primitives.ts`](examples/advanced_primitives.ts) - `BigInt`, `Symbol` registry, and `RegExp` literals.
- [`examples/async_generators.ts`](examples/async_generators.ts) - Generator functions, async generators, and `for await..of` loops.
- [`examples/node_apis.ts`](examples/node_apis.ts) - Built-in Node.js APIs (`path`, `crypto`, `fs`, `os`).

---

## Toolchain & Cross-Compilation

**ScriptGo** uses a Clang-compatible C/LLVM compiler driver to compile emitted LLVM IR and link against the lightweight runtime:

- **System Clang**: Defaults to `clang` in your `$PATH`.
- **Zig CC (`zig cc`)**: If `clang` is not installed or when compiling for cross targets (including `--target wasm32-wasi`), **ScriptGo** automatically utilizes `zig` in `$PATH` for zero-dependency builds and cross-compilation across platforms.

### Configuring Compiler Driver & Target Triple

You can configure the C compiler driver and target triple via CLI flags or environment variables:

```sh
# 1. WebAssembly / WASI compilation
scriptgo build examples/hello.ts --target wasm32-wasi -o hello.wasm

# 2. Cross-compilation for Linux x86_64
scriptgo build examples/hello.ts --cc zigcc --target x86_64-linux-gnu -o hello-linux

# 3. Via environment variables
export SCRIPTGO_CC="zigcc"
export SCRIPTGO_TARGET="wasm32-wasi"
scriptgo build examples/hello.ts -o hello.wasm
```

### Direct Cross-Compilation via `zig cc`

You can also emit LLVM IR and cross-compile with `zig cc` directly:

```sh
# 1. Emit LLVM IR
scriptgo emit examples/hello.ts -o module.ll

# 2. Cross-compile for Linux x86_64
zig cc -target x86_64-linux-gnu -O2 -x ir module.ll -x c internal/runtime/runtime.c -o hello-linux

# 3. Cross-compile for Windows x86_64
zig cc -target x86_64-windows-gnu -O2 -x ir module.ll -x c internal/runtime/runtime.c -o hello-windows.exe

# 4. Cross-compile for macOS ARM64
zig cc -target aarch64-macos -O2 -x ir module.ll -x c internal/runtime/runtime.c -o hello-macos
```

---

## Development & Contributing

We welcome contributions! Please check out [`CONTRIBUTING.md`](CONTRIBUTING.md) to get started.

```sh
# Build binary
make build

# Run all unit & integration tests
make test

# Run TypeScript-Go frontend tests
make test-frontend

# Run Node.js parity comparison benchmark
make test-parity
```

---

## Community & Security

- **Contributing Guide**: [`CONTRIBUTING.md`](CONTRIBUTING.md)
- **Code of Conduct**: [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md)
- **Security Policy**: [`SECURITY.md`](SECURITY.md)

---

## License

This project is licensed under the [MIT License](LICENSE).
