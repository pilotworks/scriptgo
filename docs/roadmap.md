# scriptgo Roadmap

## Product Direction

`scriptgo` compiles a deliberately small, strict subset of TypeScript into a
native executable. TypeScript-Go remains responsible for parsing, binding,
module resolution, and type checking. `scriptgo` owns the native subset gate,
typed IR, runtime ABI, lowering, and native backends.

The first release targets synchronous local programs on macOS ARM64 with LLVM
and Clang. Full JavaScript and npm compatibility are out of scope for the MVP.

## Current Baseline

- [x] CLI accepts one `.ts` entry point and an optional `-o` output path.
- [x] TypeScript-Go parses, resolves local `.ts` imports, and type-checks the reachable source graph.
- [x] Syntax and semantic diagnostics include the source path, offset, TypeScript code, and message.
- [x] Typed IR types, constants, arithmetic, calls, returns, printing, and a verifier exist for the MVP subset.
- [x] MVP lowering, reference interpreter, native ABI calls, LLVM IR emission, Clang executable output, and end-to-end tests exist.
- [ ] Exceptions, async code, npm/package resolution, and full JavaScript compatibility remain outside the MVP.

## Milestones

### Milestone 1: Checked Front-End Contract

Make the frontend return a stable, checked compilation unit instead of only a
statement count.

- [x] Pin the compatible TypeScript-Go revision in `internal/typescriptgo/go.mod`.
- [x] Add source files, compiler options, symbols/types, module references, and source spans to the adapter contract.
- [x] Run binding, local module resolution, and type checking for the entry point.
- [x] Define a native subset feature matrix and reject unsupported constructs with actionable diagnostics.
- [x] Add tests for valid programs, type errors, and local imports.

Acceptance: `scriptgo check entry.ts` reports deterministic parse/type/subset
diagnostics and succeeds only for a closed, supported program graph.

Verification: `go test ./...`, plus CLI tests covering valid input, missing
files, invalid extensions, type errors, and unsupported syntax.

Dependencies: current frontend adapter. Estimated scope: Medium.

### Milestone 2: Typed IR and Verifier

Replace the placeholder module with a backend-independent IR that preserves
types, evaluation order, and source locations.

- [x] Define value representations, constants, locals, arithmetic, calls, and returns for the MVP.
- [x] Attach source spans and useful names to instructions and functions.
- [x] Lower literals, local declarations, arithmetic, functions, and `return`.
- [x] Implement initial IR verification for values, instruction kinds, and return consistency.
- [x] Add golden source-to-IR tests for the MVP examples.
- [x] Add a small interpreter for semantic tests before native code generation.

Acceptance: the MVP program `add(20, 22)` lowers to verified IR and produces
the expected result in the interpreter.

Verification: golden IR tests, verifier negative tests, and differential tests
against the TypeScript/JavaScript reference runtime.

Dependencies: Milestone 1. Estimated scope: Large; implement as smaller
vertical slices for expressions, functions, and control flow.

### Checkpoint A: Frontend-to-IR

- [x] A supported synchronous program parses, type-checks, lowers, and verifies.
- [x] Unsupported features fail before backend generation.
- [x] Diagnostics retain original TypeScript locations.
- [x] `go test ./...` and `go build ./cmd/scriptgo` pass.

### Milestone 3: Runtime ABI and LLVM MVP

Turn verified IR into a native executable for the initial primitive subset.

- [x] Specify linked-runtime ABI v1 for primitive values, including conventions, `number` as `f64`, string representation, ownership, errors, compatibility, and process startup policy.
- [x] Document the temporary MVP host C ABI boundary in `internal/runtime/abi/README.md`.
- [x] Implement reference interpreter semantics plus native `printf`/`puts` calls for numeric output, strings, and `console.log`.
- [x] Generate LLVM IR for MVP constants, locals, arithmetic, calls, returns, and print calls.
- [x] Invoke Clang for the selected host target and link against the host C runtime.
- [x] Add `--emit run`, `--emit llvm-ir`, and `--emit exe` CLI modes.
- [x] Emit compiler errors for missing toolchains and unsupported target/backend combinations.

Acceptance: compiling and running the MVP program produces `42` as a native
executable, with deterministic output and exit status.

Verification: end-to-end executable tests, LLVM IR golden tests, ABI tests, and
reference-runtime differential tests for numeric and string behavior.

Dependencies: Milestones 1-2. Estimated scope: Large; land runtime ABI before
expanding code generation.

ABI v1 is intentionally primitive-only. Objects, `null`, `undefined`,
allocation beyond the documented array runtime, and runtime-managed errors
require subsequent ABI work and must not be enabled by the native subset until
their layouts and ownership rules are specified.

### Milestone 4: Modules, Arrays, and Static Objects

Expand the useful strict subset while keeping runtime behavior explicit.

- [x] Resolve and compile local TypeScript modules with deterministic initialization order.
- [x] Add runtime-managed number arrays, indexing, and bounds checks.
- [x] Add simple classes/object shapes with documented static-layout rules.
- [x] Define null/undefined representation and truthiness conversions.
- [x] Add integration fixtures for multi-file programs and object/array behavior.

Acceptance: a multi-file synchronous program using primitives, arrays, and
simple classes compiles and matches the reference interpreter on supported cases.

Verification: module-graph tests, bounds/null behavior tests, and end-to-end
differential tests.

Dependencies: Milestone 3. Estimated scope: Large.

### Checkpoint B: Native MVP Release

- [x] One-command build produces a runnable host executable.
- [x] The supported-language subset is documented and enforced.
- [x] Runtime ABI version and target assumptions are explicit.
- [x] Diagnostics and test fixtures cover every supported MVP construct.
- [x] Benchmark and smoke-test results are recorded.

### Milestone 5: Errors, Debugging, and Tooling

Make failures and intermediate artifacts usable during compiler development.

- [x] Lower unsupported constructs into source-anchored diagnostics.
- [x] Add IR dumps, LLVM dumps, and diagnostic verbosity.
- [x] Add reproducible build metadata.
- [x] Preserve source locations in LLVM debug metadata where feasible.
- [x] Add sanitizer, leak, and runtime failure test modes.
- [x] Document compiler flags, target selection, and runtime compatibility.

Acceptance: a user can identify the original TypeScript construct behind a
frontend, lowering, runtime, or backend failure.

Dependencies: Milestone 3. Estimated scope: Medium.

### Milestone 6: Deferred C Backend and Language Expansion

Only after LLVM and runtime semantics are stable, add portability and larger
language features.

- [ ] Generate C from the same Typed IR and runtime ABI.
- [ ] Prove backend parity using the shared semantic test suite.
- [ ] Add explicit FFI declarations and selected standard-library modules.
- [ ] Decide and implement memory management for object-heavy programs.
- [ ] Add exceptions and async state-machine lowering only with explicit semantics tests.

Acceptance: C and LLVM backends agree on observable behavior for the shared
subset and the expanded feature set has documented runtime costs.

Dependencies: Checkpoint B and Milestone 5. Estimated scope: XL; split each
backend/runtime feature into separate implementation tasks.

### Standard Library Compatibility Slice

The standard library follows the Node.js-compatible policy in
[`docs/stdlib.md`](stdlib.md). Implement it in this order:

1. Define a versioned built-in module manifest and canonical `node:` names.
2. Promote pure synchronous modules, starting with `node:path`.
3. Add deterministic Node-reference, interpreter, and native differential tests.
4. Add process/environment and filesystem APIs only after startup, ownership,
   encoding, and failure behavior are documented in the runtime ABI.
5. Defer callbacks, streams, networking, crypto, child processes, and worker
   APIs until object, async, and process-model semantics are complete.

## Recommended Execution Order

1. Frontend checked-program contract and feature matrix.
2. Primitive expression lowering plus IR verifier.
3. Functions/control flow plus interpreter.
4. Runtime ABI and LLVM code generation for the MVP.
5. CLI emit/build modes and end-to-end executable tests.
6. Modules, arrays, objects, and debugging artifacts.
7. C backend, FFI, exceptions, and async features.

Each item should leave the repository passing `go test ./...` and building with
`go build ./cmd/scriptgo`.

## Risks and Decisions

| Risk | Impact | Mitigation |
| --- | --- | --- |
| TypeScript semantics are more dynamic than native layouts | High | Enforce a strict subset gate and make runtime operations explicit. |
| `number` differs from native integers | High | Use `f64` for ordinary `number`; add integer types only with explicit semantics. |
| Runtime ownership is underspecified | High | Freeze the ABI and memory strategy before object/array expansion. |
| LLVM and C behavior diverge | Medium | Share Typed IR, ABI, runtime, and differential tests. |
| TypeScript-Go APIs change | Medium | Pin the revision and isolate all dependency access in the adapter. |
| npm compatibility expands scope | High | Support only closed local modules and documented runtime modules initially. |

## Open Questions

- Should `check`, `lower`, and `build` be separate subcommands, or should the current single-entry CLI remain the compatibility surface?
- Is Clang/LLVM 18 available in every supported development and CI environment?
- Which memory strategy is acceptable for the first object runtime: ownership, reference counting, or garbage collection?
- Which strict TypeScript constructs are required for the first real user program beyond the documented `add` example?

## Current Target Assumptions

- Development and MVP smoke tests target macOS ARM64 (`darwin/arm64`).
- Native builds use the `clang` executable available in `PATH`; the current
  verification host uses Apple Clang 21.
- LLVM output uses opaque pointers and the host C ABI; target selection and
  cross-compilation are exposed through `--target`, while `native` remains the
  default host target.
- `--debug` emits stable LLVM debug metadata using source basenames and a
  reproducible compilation directory. `--sanitize` passes a comma-separated
  sanitizer list to Clang.
- The standard smoke fixture is `console.log(20 + 22)`, and the native build
  benchmark is `BenchmarkBuildNative` in `internal/compiler/compiler_test.go`.
- Run `go test ./...`, `go build ./cmd/scriptgo`, and
  `go test ./internal/compiler -run '^$' -bench BenchmarkBuildNative -benchtime=1x`
  to repeat the baseline checks.
