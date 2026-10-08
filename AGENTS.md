# AGENTS.md

These rules apply to the whole repository. Keep this file focused on
`scriptgo` decisions; upstream documentation remains the source of truth for
general TypeScript-Go and LLVM behavior.

## Source Of Truth

- Use [TypeScript](https://github.com/microsoft/TypeScript) (TypeScript in Go) for parsing,
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
- Always choose solutions that are sustainable, scalable, and aligned with the
  long-term architecture. Avoid short-term workarounds, special cases, and
  patches that shift complexity into a later stage; if a temporary workaround
  is unavoidable, document its scope, trade-offs, removal criteria, and
  follow-up path.
- Treat the module ownership rules below as hard boundaries. Do not move logic
  across a boundary merely because the destination package is convenient.
  Always perform a module responsibility review before and after adding features
  to ensure no package violates its boundaries or takes on forbidden responsibilities.
- Keep the frontend-to-IR boundary independent from any one native backend.
- Make JavaScript/TypeScript semantic differences explicit in the native subset
  policy, runtime ABI, diagnostics, and tests.
- Add source spans to diagnostics and IR instructions whenever a new lowering
  boundary can report an error.
- Keep each Go file focused on one responsibility. Split a file by stage or
  concern before adding more code when it exceeds roughly 500 lines; a file
  over 700 lines requires an explicit reason in the change description.
- Do not create catch-all `helpers`, `utils`, or `common` packages. Put a helper
  beside the boundary that owns its behavior, or document why it is genuinely
  cross-cutting.
- Always add corpus test cases under `internal/compiler/testdata/corpus/`
  (`main.ts` with `run.expected`, `run.err`, or `check.err`) for every newly implemented
  feature, syntax construct, operator, or standard library function.
- Always keep [`docs/typescript-parity-report.md`](docs/typescript-parity-report.md)
  strictly synchronized with the codebase. Whenever language features, syntax
  constructs, standard library APIs, diagnostics, or runtime capabilities are
  implemented, modified, or verified, update the parity report tables and test
  metrics accordingly.
- Keep markdown documentation concise, complete, and actionable. Document only
  essential context, architectural decisions, contracts, and concrete behavior;
  avoid unnecessary filler, boilerplate, and redundant explanations.
- Always run full regression tests and build verification on every change:
  run `go test -count=1 ./...`, `go test -count=1 ./internal/typescriptgo/...`,
  and `go build ./cmd/scriptgo`.
- Keep the current CLI contract stable unless a documented roadmap item requires
  a breaking change.

## Module Ownership And Forbidden Responsibilities

These responsibilities are strict. If a feature appears to belong to two
modules, define the contract at the boundary and keep the implementation in the
module that owns the behavior.

| Module | May own | Must not own |
| --- | --- | --- |
| `cmd/scriptgo` | Flag parsing, input/output paths, process exit codes, user-facing errors | TypeScript semantics, AST traversal, lowering, IR construction, LLVM text, runtime behavior |
| `internal/compiler` | Stage orchestration, artifact selection, temporary toolchain files, invoking Clang | Parsing, type checking, feature semantics, instruction emission |
| `internal/typescriptgo` | The pinned TypeScript-Go adapter and conversion of upstream results into stable adapter data | Native subset policy, IR, LLVM, runtime ABI, a replacement parser/type system |
| `internal/frontend` | Checked program creation, reachable module graph, normalized source data, frontend diagnostics, the frontend-to-lowering contract (`contract.go`) | Backend selection, native layout, runtime calls, LLVM emission, execution |
| `internal/lowering` | Native subset validation and conversion of checked source data into backend-independent IR | Direct LLVM/C emission, native process startup, TypeScript-Go compiler policy |
| `internal/ir` | Types, instructions, modules, source metadata, verification invariants | TypeScript AST/API usage, backend-specific syntax, runtime implementation, CLI policy |
| `internal/opt` | Target-independent Typed IR optimization passes, dead code elimination, constant folding, CSE, LICM | TypeScript AST/API usage, backend-specific LLVM emission, runtime ABI or C source implementation |
| `internal/runtime` | The native ABI contract and linked runtime services for values, ownership, startup, and errors | TypeScript parsing/type checking, lowering policy, backend orchestration |
| `internal/backend/llvm` | Translation of verified IR into LLVM IR and LLVM target details | TypeScript AST inspection, type checking, subset decisions |
| `internal/pkgmgr` | npm-compatible manifests, resolution, lockfiles, content store, installs, `package.json` tasks | Compilation stages, TypeScript semantics, runtime ABI |
| `internal/audit`, `internal/spec`, `cmd/parity` | Developer tooling: Node.js API catalog, corpus coverage audit, parity runner | Anything linked into the `scriptgo` CLI or compiled programs |
| `internal/test262`, `cmd/test262` | Developer tooling: test262 metadata, the native-subset harness, conformance runner and report | Compiler stages (it builds through an injected `Builder`), vendoring the suite |

Dependency direction must remain acyclic:

```text
cmd/scriptgo -> compiler, pkgmgr
compiler -> frontend -> typescriptgo -> TypeScript-Go
compiler -> lowering -> frontend (contract types only), ir
compiler -> opt -> ir
compiler -> backend/llvm -> ir
compiler -> runtime
cmd/parity -> audit -> frontend, spec
cmd/test262 -> compiler, test262
```

Only `internal/frontend` imports the TypeScript-Go adapter; later stages use
the aliases in `internal/frontend/contract.go`. `internal/runtime` provides the
native ABI contract and linked C runtime source. `TestDependencyDirection` in
`internal/compiler` enforces this graph; a new package must be added there and
here before it can be imported.

Compiler output must be deterministic: never let Go map iteration order decide
emitted order or a "first match" lookup. Iterate sorted keys
(`slices.Sorted(maps.Keys(m))`) or keep an explicit order slice.
`TestCompileIsDeterministicAcrossRuns` guards this.

## File Boundaries And Splitting Rules

- Keep orchestration, data models, conversion logic, diagnostics, and emission
  in separate files when they grow independently.
- Split `internal/typescriptgo` by adapter model, parsing, program checking,
  syntax normalization, and diagnostics rather than growing one `parse.go`.
- Split `internal/lowering` by expressions, statements, calls/runtime
  operations, and subset diagnostics as control flow is added.
- Split `internal/backend/llvm` by module/function emission, constants/values,
  runtime ABI declarations, and debug/target metadata as those concerns grow.
- Keep tests next to the file or package that owns the behavior. A test that
  crosses multiple boundaries belongs in the owning integration package and
  must not become a reason to merge unrelated implementations into one file.
- When a change makes a file larger than the guideline, split first, then add
  the feature. Preserve public contracts while moving private helpers.

## Working With Upstream

Before changing TypeScript-facing behavior, inspect the pinned TypeScript-Go
revision in `internal/typescriptgo/go.mod` and its public APIs. Adapt upstream
results in `internal/frontend`; do not fork or restate upstream implementation
rules in `scriptgo`.

Before changing LLVM emission, check the relevant LLVM reference documentation
and keep target-specific assumptions in `internal/backend/llvm` and the runtime
ABI documentation.

The upstream TypeScript contribution and automation policies are documented
in its [CONTRIBUTING.md](https://github.com/microsoft/TypeScript/blob/main/CONTRIBUTING.md).
This repository adopts those policies by reference where applicable; it does
not copy them here.
