# scriptgo

`scriptgo` is a Go-based native compiler that aims to run TypeScript and
JavaScript with Node.js-compatible semantics while compiling eligible code to
native executables. The current implementation is a deliberately small,
synchronous subset; it reuses TypeScript-Go for the language front end and LLVM
for the native middle/backend path.

## Documentation

- [`docs/application-structure.md`](docs/application-structure.md) - repository
  layout, ownership, and dependency direction.
- [`docs/typescript-to-native.md`](docs/typescript-to-native.md) - compiler
  architecture, pipeline boundaries, and native subset policy.
- [`docs/compilation-tiers.md`](docs/compilation-tiers.md) - Static, Dynamic,
  and Unsupported compilation policy.
- [`docs/roadmap.md`](docs/roadmap.md) - milestones and acceptance criteria.
- [`docs/stdlib.md`](docs/stdlib.md) - Node.js-compatible standard-library policy and scope.

The documentation follows the structure of the [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html):
start with the project model and basic concepts, then use architecture and
reference material for deeper details. Native code generation follows the
modular, reusable compiler infrastructure model described by [LLVM](https://llvm.org/).

## Status

The repository now has a working synchronous MVP pipeline. TypeScript-Go parses,
resolves local modules, and type-checks the reachable graph; scriptgo lowers a
small subset into verified typed IR, can interpret it, and can emit LLVM IR or
an executable through Clang. Dense `number[]` arrays and static number/string
class fields are supported; `null`, `undefined`, exceptions, async code, npm
resolution, and full JavaScript compatibility are deferred from the MVP, not
the long-term compatibility direction.

Compilation follows three explicit tiers: Static is the default and compiles
eligible code directly to native code; Dynamic is planned behind an explicit
`--dynamic` flag and will use embedded QuickJS-ng for eligible JavaScript/npm
code and `any`; Unsupported constructs produce a compile error. The current
MVP is Static-only, so it does not yet accept `--dynamic`.

## CLI Usage

```sh
# Run a TypeScript program (interpreter or native)
scriptgo run examples/hello.ts
scriptgo run --native examples/hello.ts

# Build a standalone native binary
scriptgo build examples/hello.ts -o hello

# Type-check and validate program subset
scriptgo check examples/hello.ts

# Emit intermediate representations (LLVM IR or Typed IR)
scriptgo emit examples/hello.ts
scriptgo emit examples/hello.ts --mode typed-ir -o out.ir
```


Native builds accept `-target native` (or a Clang target triple), `-debug` for
source debug metadata, and `-sanitize address,undefined,leak` for opt-in Clang
sanitizers. Generated LLVM includes the compiler version, runtime ABI version,
target, and a SHA-256 hash of the reachable local source graph.

## Development

```sh
go test ./...
go build ./cmd/scriptgo
```

The first implementation keeps the frontend and backend boundaries in
`internal/frontend`, `internal/ir`, and `internal/compiler` so they can be
replaced without changing the CLI contract.
