# scriptgo

<div align="center">

[![CI](https://github.com/pilotworks/scriptgo/actions/workflows/ci.yml/badge.svg)](https://github.com/pilotworks/scriptgo/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org)
[![TypeScript Parity](https://img.shields.io/badge/TypeScript%20Parity-100%25%20(136%2F136)-success.svg)](docs/typescript-parity-report.md)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platforms](https://img.shields.io/badge/Platforms-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey.svg)](#toolchain--cross-compilation)

**A high-performance Ahead-Of-Time (AOT) compiler compiling TypeScript to native standalone executables with Node.js parity.**

</div>

---

`scriptgo` is a native compiler that runs TypeScript and JavaScript with Node.js-compatible semantics while compiling eligible code directly to standalone native binaries. It combines the official [TypeScript-Go](https://github.com/microsoft/typescript-go) compiler frontend for parsing, type-checking, and diagnostics with an independent **Typed IR** system and an **LLVM IR / Native Machine Code** backend.

---

## Highlights & Features

- **High-Performance AOT Compilation**: Compiles TypeScript directly to native machine code via LLVM without requiring a JavaScript virtual machine for static code.
- **Node.js Semantic Parity**: 100% pass rate across the 136-case regression test corpus checked against Node.js v22+.
- **Zero-Dependency Native Builds**: Automatically uses system `clang` or auto-detects `zig cc` for hassle-free out-of-the-box compilation and seamless cross-compilation (macOS, Linux, Windows).
- **Dual Execution Modes**: Fast reference interpreter for instant execution and native compilation for optimized binary builds.
- **Comprehensive TypeScript Support**:
  - **Types & Primitives**: `number` (IEEE-754), `bigint`, `string` (UTF-8), `boolean`, `symbol` (with Symbol Registry), `null`, `undefined`, `unknown` (with type narrowing), Tuples, Enums (numeric, string, reverse mappings), Union types (`T | null | undefined`), Monomorphized Generics.
  - **Control Flow**: `if`/`else`, `switch`/`case` (with fallthrough), `while`, `do..while`, `for`, `for..of`, `for..in`, `for await..of`, Labeled statements (`break label`, `continue label`), `try`/`catch`/`finally`, `throw`, Array & Object destructuring, Spread/Rest (`...`), Tagged template literals, Optional chaining & calls (`?.`, `fn?.()`).
  - **Functions & Closures**: Lexical closures, arrow functions, default/optional/rest parameters, Generators (`function*`, `yield`, `yield*`), Async Generators.
  - **OOP & Classes**: Constructors, properties, static fields/methods, Class Static Blocks (`static { ... }`), Getters/Setters, Inheritance (`extends`, `super`), Polymorphic VTables, `instanceof`.
  - **Async Runtime**: `Promise` (resolve, reject, chaining), `async`/`await`, microtask queue execution conforming to JavaScript event loop ordering.
  - **Standard Library & Node.js APIs**: `console`, `Math`, `String`, `Array` (with `map`, `filter`, `reduce`, `find`, `some`, `every`), `Date`, `JSON`, `RegExp` (literals and methods), `BigInt`, `Symbol`, `node:fs`, `node:path`, `node:os`, `node:process`, `node:crypto`, `performance.now()`, Base64 (`btoa`, `atob`, `Buffer.from`).

---

## Documentation

- [`docs/typescript-parity-report.md`](docs/typescript-parity-report.md) - Comprehensive TypeScript/Node.js feature matrix and parity test report.
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
# Run using the built-in reference interpreter
scriptgo run examples/hello.ts

# Compile to native binary and run immediately
scriptgo run --native examples/hello.ts
```

### Building Standalone Executables

```sh
# Build a native binary for the host platform
scriptgo build examples/hello.ts -o hello

# Build with debug symbols
scriptgo build examples/hello.ts --debug -o hello-debug

# Build with Clang sanitizers (address, undefined, leak)
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

`scriptgo` uses a Clang-compatible C/LLVM compiler driver to compile emitted LLVM IR and link against the lightweight runtime:

- **System Clang**: Defaults to `clang` in your `$PATH`.
- **Zig CC (`zig cc`)**: If `clang` is not installed or when `--cc zigcc` is specified, `scriptgo` automatically detects `zig` in `$PATH` for zero-dependency builds and cross-compilation across platforms.

### Configuring Compiler Driver & Target Triple

You can configure the C compiler driver and target triple via CLI flags or environment variables:

```sh
# 1. Via CLI flags
scriptgo build examples/hello.ts --cc zigcc --target x86_64-linux-gnu -o hello-linux
scriptgo run --native --cc zigcc examples/hello.ts

# 2. Via environment variables
export SCRIPTGO_CC="zigcc"
export SCRIPTGO_TARGET="x86_64-linux-gnu"
scriptgo build examples/hello.ts -o hello-linux

# 3. Inline environment overrides
SCRIPTGO_CC="zigcc" SCRIPTGO_TARGET="aarch64-macos" scriptgo build examples/hello.ts -o hello-macos
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

# Run Node.js parity comparison benchmark (136/136 tests)
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
