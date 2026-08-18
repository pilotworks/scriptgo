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
- [x] TypeScript-Go adapter parses a single source file.
- [x] Syntax diagnostics include the source path, offset, TypeScript code, and message.
- [x] Initial IR types exist for `void`, `bool`, `number`, and `string`.
- [ ] Type checking, module resolution, lowering, runtime, executable output, and LLVM emission are not implemented.

## Milestones

### Milestone 1: Checked Front-End Contract

Make the frontend return a stable, checked compilation unit instead of only a
statement count.

- [ ] Pin and document the compatible TypeScript-Go revision.
- [ ] Add source files, compiler options, symbols/types, module references, and source spans to the adapter contract.
- [ ] Run binding, module resolution, and type checking for the entry point.
- [ ] Define a native subset feature matrix and reject unsupported constructs with actionable diagnostics.
- [ ] Add tests for valid programs, type errors, local imports, and source locations.

Acceptance: `scriptgo check entry.ts` reports deterministic parse/type/subset
diagnostics and succeeds only for a closed, supported program graph.

Verification: `go test ./...`, plus CLI tests covering valid input, missing
files, invalid extensions, type errors, and unsupported syntax.

Dependencies: current frontend adapter. Estimated scope: Medium.

### Milestone 2: Typed IR and Verifier

Replace the placeholder module with a backend-independent IR that preserves
types, evaluation order, and source locations.

- [ ] Define value representations, constants, locals, arithmetic, comparisons, calls, branches, loops, and returns.
- [ ] Attach source spans and useful names to instructions and functions.
- [ ] Lower literals, local declarations, assignments, arithmetic, functions, and `if`/`return`.
- [ ] Implement IR verification for type consistency, control-flow validity, and terminated blocks.
- [ ] Add golden source-to-IR tests for the MVP examples.
- [ ] Add a small interpreter or execution harness for semantic tests before native code generation.

Acceptance: the MVP program `add(20, 22)` lowers to verified IR and produces
the expected result in the interpreter.

Verification: golden IR tests, verifier negative tests, and differential tests
against the TypeScript/JavaScript reference runtime.

Dependencies: Milestone 1. Estimated scope: Large; implement as smaller
vertical slices for expressions, functions, and control flow.

### Checkpoint A: Frontend-to-IR

- [ ] A supported synchronous program parses, type-checks, lowers, and verifies.
- [ ] Unsupported features fail before backend generation.
- [ ] Diagnostics retain original TypeScript locations.
- [ ] `go test ./...` and `go build ./cmd/scriptgo` pass.

### Milestone 3: Runtime ABI and LLVM MVP

Turn verified IR into a native executable for the initial primitive subset.

- [ ] Specify ABI conventions, `number` as `f64`, string representation, ownership, errors, and process startup.
- [ ] Implement runtime startup, numeric operations, strings, and `console.log`.
- [ ] Generate LLVM IR for constants, locals, arithmetic, calls, branches, returns, and runtime calls.
- [ ] Invoke Clang for the selected host target and link the runtime.
- [ ] Add `--backend llvm`, `--emit typed-ir`, and `--emit llvm-ir` CLI modes.
- [ ] Emit compiler errors for missing toolchains and unsupported target/backend combinations.

Acceptance: compiling and running the MVP program produces `42` as a native
executable, with deterministic output and exit status.

Verification: end-to-end executable tests, LLVM IR golden tests, ABI tests, and
reference-runtime differential tests for numeric and string behavior.

Dependencies: Milestones 1-2. Estimated scope: Large; land runtime ABI before
expanding code generation.

### Milestone 4: Modules, Arrays, and Static Objects

Expand the useful strict subset while keeping runtime behavior explicit.

- [ ] Resolve and compile local TypeScript modules with deterministic initialization order.
- [ ] Add runtime-managed arrays, indexing, bounds checks, and common array operations.
- [ ] Add simple classes/object shapes with documented static-layout rules.
- [ ] Define null/undefined representation and truthiness conversions.
- [ ] Add integration fixtures for multi-file programs and object/array behavior.

Acceptance: a multi-file synchronous program using primitives, arrays, and
simple classes compiles and matches the reference runtime on supported cases.

Verification: module-graph tests, bounds/null behavior tests, and end-to-end
differential tests.

Dependencies: Milestone 3. Estimated scope: Large.

### Checkpoint B: Native MVP Release

- [ ] One-command build produces a runnable host executable.
- [ ] The supported-language subset is documented and enforced.
- [ ] Runtime ABI version and target assumptions are explicit.
- [ ] Diagnostics and test fixtures cover every supported MVP construct.
- [ ] Benchmark and smoke-test results are recorded.

### Milestone 5: Errors, Debugging, and Tooling

Make failures and intermediate artifacts usable during compiler development.

- [ ] Lower unsupported constructs into source-anchored diagnostics.
- [ ] Add IR dumps, LLVM dumps, diagnostic verbosity, and reproducible build metadata.
- [ ] Preserve source locations in LLVM debug metadata where feasible.
- [ ] Add sanitizer, leak, and runtime failure test modes.
- [ ] Document compiler flags, target selection, and runtime compatibility.

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
