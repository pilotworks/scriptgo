# Application Structure

`scriptgo` is a Go compiler application with a TypeScript-Go frontend, a
backend-independent typed IR, a reference interpreter, and an LLVM native
backend. This document separates the repository as it exists today from the
future layout used by the roadmap. A `planned` directory is an architectural
boundary only; it must not be created until the related implementation slice is
ready.

## Design Principles

- Keep the CLI thin. It selects inputs and reports process-level errors; it
  does not implement TypeScript semantics.
- Reuse TypeScript-Go for parsing, binding, module resolution, type checking,
  and diagnostics. The local adapter isolates that dependency.
- Keep typed IR independent of LLVM and any other native backend.
- Make runtime representation, ownership, ABI, and target assumptions explicit.
- Keep tests next to the package that owns the behavior they verify; use shared
  fixtures only when a test crosses a pipeline boundary.
- Prefer additive, vertical slices that leave `go test ./...` and the CLI build
  working after each change.

## Current Repository Layout

This is the structure currently present in the repository:

```text
.
├── AGENTS.md
├── README.md
├── go.mod                       # Root module: github.com/pilotworks/scriptgo
├── go.sum
├── go.work                      # Workspace including the local adapter module
├── go.work.sum
├── cmd/
│   └── scriptgo/
│       └── main.go              # CLI entry point and -o output flag
├── internal/
│   ├── compiler/
│   │   ├── compiler.go          # Pipeline orchestration and build modes
│   │   └── compiler_test.go     # Compiler-level behavior tests
│   ├── frontend/
│   │   └── source.go            # Program validation and frontend normalization
│   ├── lowering/
│   │   ├── lowering.go          # Checked TypeScript subset -> typed IR
│   │   └── lowering_test.go     # Lowering and subset tests
│   ├── ir/
│   │   └── ir.go                # Initial typed IR data contract
│   ├── interpreter/
│   │   ├── interpreter.go        # Reference execution engine for typed IR
│   │   └── interpreter_test.go   # Interpreter semantic tests
│   ├── backend/
│   │   └── llvm/
│   │       ├── emit.go           # Typed IR -> LLVM IR
│   │       └── emit_test.go      # LLVM emission tests
│   ├── runtime/
│   │   └── abi.md                # Current host ABI contract; no interpreter code
│   └── typescriptgo/
│       ├── go.mod               # Separate adapter module
│       ├── go.sum
│       └── parse.go             # Small wrapper around TypeScript-Go parsing
├── examples/
│   └── hello.ts                 # Minimal CLI input fixture
└── docs/
    ├── application-structure.md
    ├── roadmap.md
    └── typescript-to-native.md
```

### Current execution path

```text
cmd/scriptgo/main.go
    -> internal/compiler.Compile
        -> internal/frontend.NewProgram
        -> internal/typescriptgo.Check
            -> github.com/microsoft/typescript-go
        -> internal/ir.Module
        -> internal/lowering.Lower
        -> internal/backend/llvm.Emit or internal/interpreter.Execute
    -> stdout or -o output file
```

The current implementation checks the reachable local module graph, lowers the
supported synchronous subset, and emits LLVM IR or runs the reference
interpreter. The native ABI is currently provided by host `printf`/`puts` and
Clang; managed strings, arrays, objects, and exceptions remain planned.

## Target Repository Layout

The following layout is the intended destination described by the roadmap:

```text
internal/
├── compiler/
│   ├── compiler.go              # Stage orchestration
│   ├── options.go               # Planned: normalized compiler options
│   └── diagnostics.go           # Planned: cross-stage diagnostics
├── typescriptgo/                # Existing adapter module
├── frontend/
│   ├── source.go                # Existing program model
│   ├── project.go               # Planned: module graph and program creation
│   └── diagnostics.go           # Planned: source-anchored errors
├── interpreter/                 # Existing: reference execution; never linked natively
│   ├── interpreter.go            # Typed IR execution engine
│   └── interpreter_test.go       # Semantic oracle tests
├── lowering/                    # Existing: checked TypeScript -> typed IR
│   ├── lowering.go              # Existing MVP lowering
│   ├── expressions.go           # Planned split: literals, operators, calls
│   ├── statements.go            # Planned split: declarations, branches, returns
│   └── subset.go                # Planned: supported-feature gate
├── ir/
│   ├── ir.go                    # Existing module/type/instruction model
│   ├── verify.go                # Planned: IR validity checks
│   └── dump.go                  # Planned: stable human-readable IR output
├── runtime/                     # Existing ABI contract; implementation planned
│   ├── abi.md                   # Host ABI plus frozen primitive linked ABI v1
│   ├── startup/                 # Planned: process startup and exit handling
│   └── values/                  # Planned: strings, arrays, objects, errors
└── backend/
    └── llvm/                    # Existing MVP: typed IR -> LLVM IR
        ├── emit.go              # Existing module/function emission
        ├── target.go            # Planned: target triple and data layout
        └── debug.go             # Planned: source/debug metadata
```

Cross-stage fixtures will be added only when needed:

```text
testdata/                       # Planned: shared fixtures and goldens
├── frontend/                   # Planned: diagnostics and module graphs
├── ir/                         # Planned: source-to-IR fixtures
└── llvm/                       # Planned: LLVM output fixtures
```

Do not add empty placeholder directories. Add a package with its first
behavior, focused tests, and a roadmap slice that explains the boundary.

## Ownership

| Area | Owns | Must not own |
| --- | --- | --- |
| `cmd/scriptgo` | CLI flags, input selection, output writing, exit codes | Parsing, type checking, lowering, LLVM details |
| `internal/compiler` | Pipeline orchestration, options, artifact selection | Becoming a miscellaneous utility package |
| `internal/typescriptgo` | Versioned dependency isolation and adapter helpers | A second parser or type system |
| `internal/frontend` | Program creation, module graph, checked input, source spans | Native ABI, runtime calls, LLVM selection |
| `internal/lowering` | Native subset checks and explicit conversion/runtime operations | Backend-specific emission or CLI behavior |
| `internal/ir` | Backend-independent types, values, instructions, blocks, spans, verifier | TypeScript-Go internals or LLVM APIs |
| `internal/interpreter` | Reference execution and semantic oracle tests | Native executable startup or ABI implementation |
| `internal/runtime` | Native ABI contract and future value/startup services | Reference interpretation, TypeScript syntax, frontend analysis |
| `internal/backend/llvm` | Verified IR to LLVM IR, target data, debug metadata | Reimplementing TypeScript semantics |

## Dependency Direction

```text
cmd/scriptgo
    -> internal/compiler
        -> internal/frontend -> internal/typescriptgo -> TypeScript-Go
        -> internal/lowering -> internal/ir
        -> internal/interpreter (reference execution only)
        -> internal/backend/llvm -> internal/ir
```

The IR is the contract between language-facing analysis and native backends.
Backends consume verified IR; they do not inspect TypeScript ASTs. Runtime and
ABI decisions are explicit inputs to lowering and linking.

The current LLVM MVP calls the host C ABI (`printf` and `puts`) directly, so
`internal/runtime` remains documentation-only while linked ABI v1 is being
specified. ABI v1 is primitive-only; managed values or startup services require
a later implementation slice and must not be added to lowering before their
representation and ownership contract is tested. The package must not absorb
the reference interpreter or become a general utility package.

Rules for imports:

1. `cmd` imports the compiler package only.
2. The compiler coordinates stages; stages do not call back into the CLI.
3. Frontend code may use TypeScript-Go through the adapter, but IR and backend
   code must not depend on TypeScript-Go APIs.
4. Lowering produces IR; it does not emit LLVM IR directly.
5. LLVM consumes verified IR; it does not inspect TypeScript ASTs or diagnostics.
6. Runtime and ABI definitions are consumed by lowering and backend integration,
   not by frontend parsing.
7. Shared helpers belong in the owning package first. Create a package only for
   a stable cross-boundary responsibility.

## Data And Artifact Boundaries

| Boundary | Input | Output | Failure policy |
| --- | --- | --- | --- |
| CLI -> compiler | Entry path, flags, output path | Compilation result or process error | Report actionable error and exit non-zero |
| TypeScript-Go -> frontend | Source files, options, diagnostics | Normalized checked program | Preserve original source locations |
| Frontend -> lowering | Closed module graph and semantic data | Native compilation input | Reject graph/type errors before lowering |
| Lowering -> IR | Checked data and subset policy | Typed IR module | Reject unsupported or ambiguous semantics |
| IR -> backend | Verified module and target options | LLVM IR or future backend artifact | Never emit from invalid IR |
| Backend/toolchain -> user | IR, runtime, target | Object files/native executable | Surface toolchain/linker failures clearly |

Intermediate artifacts have explicit ownership:

- `typed-ir`: backend-independent debugging and golden-test output;
- `llvm-ir`: LLVM backend output before toolchain invocation;
- executable/object files: build outputs, never source-controlled;
- diagnostics: source-anchored compiler output whenever a TypeScript span exists.

## Tests And Fixtures

Tests are colocated with the package that owns the behavior:

```text
internal/compiler/compiler_test.go       # Current pipeline behavior
internal/frontend/source_test.go         # Planned: normalization/diagnostics
internal/lowering/*_test.go              # Current MVP lowering tests
internal/ir/*_test.go                    # Planned: dedicated verifier tests
internal/backend/llvm/*_test.go          # Current LLVM emission tests
internal/interpreter/*_test.go           # Current semantic oracle tests
testdata/                                # Planned: shared cross-stage inputs
```

Use `examples/` for short, readable programs intended for humans and smoke
tests. Use `testdata/` for exact output, diagnostics, or module-graph contracts.
Do not place generated LLVM, native binaries, profiles, or temporary compiler
output in the repository unless a test explicitly treats it as a checked-in
golden artifact.

## CLI To Folder Mapping

The current CLI supports one entry point and optional output:

```sh
go run ./cmd/scriptgo examples/hello.ts
go run ./cmd/scriptgo -emit run examples/hello.ts
go run ./cmd/scriptgo -emit llvm-ir -o out.ll examples/hello.ts
go run ./cmd/scriptgo -emit exe -o hello examples/hello.ts
```

Implemented modes map to pipeline boundaries rather than packages:

| Command/artifact | Boundary | Purpose |
| --- | --- | --- |
| `--emit run` | frontend -> IR -> interpreter | Execute reference semantics |
| `--emit llvm-ir` | IR -> LLVM backend | Inspect LLVM IR before toolchain invocation |
| `--emit exe` | IR -> LLVM -> Clang | Produce a native executable |
| `check`, `lower`, `--emit typed-ir` | Planned boundaries | Add when their CLI artifacts have stable contracts |

The CLI currently accepts one entry point and keeps these modes stable while
the roadmap expands the available artifacts.

## Adding A New Component

Before adding a directory or package:

1. Identify the pipeline boundary and the artifact it owns.
2. Confirm the dependency direction remains acyclic.
3. Reuse TypeScript-Go or LLVM APIs where they already provide the behavior;
   do not duplicate their implementation.
4. Add the smallest working slice and package-level tests.
5. Update this document only when the structure or ownership has changed;
   update `docs/roadmap.md` for future-only work.

## Development Flow

1. Read the relevant TypeScript Handbook topic before changing language-facing
   behavior. Reuse TypeScript-Go APIs and behavior where they already exist.
2. Extend the frontend contract only when the next pipeline stage needs data.
3. Lower one supported language construct into typed IR and verify it.
4. Add backend output only after the IR contract and semantic tests are stable.
5. Keep examples and tests close to the boundary they validate.

For LLVM-specific behavior, use the [LLVM Language Reference](https://llvm.org/docs/LangRef.html),
the [LLVM Programmer's Manual](https://llvm.org/docs/ProgrammersManual.html), and
the [LLVM Code Generator documentation](https://llvm.org/docs/CodeGenerator.html)
as the authoritative references. Do not invent an incompatible IR dialect in
project documentation.
