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

The documentation follows the structure of the [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html):
start with the project model and basic concepts, then use architecture and
reference material for deeper details. Native code generation follows the
modular, reusable compiler infrastructure model described by [LLVM](https://llvm.org/).

## Status

This repository is an executable scaffold. The CLI parses and validates a `.ts`
entry point with the TypeScript-Go adapter, then emits a textual backend stub
while lowering passes, runtime, and LLVM backend are developed.

## Run

```sh
go run ./cmd/scriptgo examples/hello.ts
go run ./cmd/scriptgo -o out.ll examples/hello.ts
```

The output is intentionally not executable LLVM yet. It marks the contract
between the frontend, typed IR, and backend.

## Development

```sh
go test ./...
go build ./cmd/scriptgo
```

The first implementation keeps the frontend and backend boundaries in
`internal/frontend`, `internal/ir`, and `internal/compiler` so they can be
replaced without changing the CLI contract.
