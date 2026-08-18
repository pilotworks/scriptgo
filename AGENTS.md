# AGENTS.md

These rules apply to the whole repository. Keep this file focused on
`scriptgo` decisions; upstream documentation remains the source of truth for
general TypeScript-Go and LLVM behavior.

## Source Of Truth

- Use [TypeScript-Go](https://github.com/microsoft/typescript-go) for parsing,
  binding, module resolution, type checking, diagnostics, and TypeScript
  semantics. Do not implement a parallel parser or type system here.
- Use the [TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)
  for progressive explanations of TypeScript concepts. Link to it instead of
  copying language reference material into this repository.
- Use [LLVM](https://llvm.org/), especially its
  [Language Reference](https://llvm.org/docs/LangRef.html), for LLVM IR and
  native toolchain behavior. Do not document a project-specific LLVM dialect.

## Repository Rules

- Preserve the package ownership and dependency direction documented in
  [`docs/application-structure.md`](docs/application-structure.md).
- Keep the frontend-to-IR boundary independent from any one native backend.
- Make JavaScript/TypeScript semantic differences explicit in the native subset
  policy, runtime ABI, diagnostics, and tests.
- Add source spans to diagnostics and IR instructions whenever a new lowering
  boundary can report an error.
- Prefer small vertical changes: update one pipeline boundary, add focused tests,
  then run `go test ./...` and `go build ./cmd/scriptgo`.
- Keep the current CLI contract stable unless a documented roadmap item requires
  a breaking change.

## Working With Upstream

Before changing TypeScript-facing behavior, inspect the pinned TypeScript-Go
revision in `internal/typescriptgo/go.mod` and its public APIs. Adapt upstream
results in `internal/frontend`; do not fork or restate upstream implementation
rules in `scriptgo`.

Before changing LLVM emission, check the relevant LLVM reference documentation
and keep target-specific assumptions in `internal/backend/llvm` and the runtime
ABI documentation.

The upstream TypeScript-Go contribution and automation policies are documented
in its [CONTRIBUTING.md](https://github.com/microsoft/typescript-go/blob/main/CONTRIBUTING.md).
This repository adopts those policies by reference where applicable; it does
not copy them here.
