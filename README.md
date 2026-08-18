# scriptgo

`scriptgo` is a Go-based native compiler for a deliberately small, typed
subset of TypeScript. It reuses TypeScript-Go for the language front end and
LLVM for the native middle/backend path.

## Documentation

- [`docs/application-structure.md`](docs/application-structure.md) - repository
  layout, ownership, and dependency direction.
- [`docs/typescript-to-native.md`](docs/typescript-to-native.md) - compiler
  architecture, pipeline boundaries, and native subset policy.
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
resolution, and full JavaScript compatibility remain outside the MVP.

## Run

```sh
go run ./cmd/scriptgo examples/hello.ts
go run ./cmd/scriptgo -emit run examples/hello.ts
go run ./cmd/scriptgo -emit llvm-ir -o out.ll examples/hello.ts
go run ./cmd/scriptgo -emit exe -o hello examples/hello.ts
```

`-emit run` uses the reference interpreter. `-emit llvm-ir` writes LLVM IR, and
`-emit exe` invokes the host `clang` to produce a native executable.

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
