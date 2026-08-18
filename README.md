# scriptgo

`scriptgo` is a Go-based native compiler for a deliberately small, typed
subset of TypeScript. The long-term pipeline is described in
[`docs/typescript-to-native.md`](docs/typescript-to-native.md).

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
