# Contributing to ScriptGo

Thank you for your interest in contributing to **ScriptGo**! We welcome contributions of all kinds: bug fixes, new language features, standard library modules, documentation improvements, and performance enhancements.

---

## 1. Prerequisites

To build and test ScriptGo locally, you need:

- **Go**: Version `1.24+` installed.
- **Node.js**: Version `v22+` installed (used as the semantic reference oracle for parity testing).
- **C/LLVM Toolchain**:
  - **Clang** (`clang` in `PATH`) **OR**
  - **Zig** (`zig` in `PATH`, which provides `zig cc`).

---

## 2. Getting Started

### Clone and Build

```sh
git clone https://github.com/pilotworks/scriptgo.git
cd scriptgo

# Build the CLI
go build ./cmd/scriptgo

# Test running a sample TypeScript file
./scriptgo run examples/hello.ts
./scriptgo run --native examples/hello.ts
```

---

## 3. Development Workflow & Commands

We provide a `Makefile` for standard development tasks:

```sh
# Run all unit and integration tests
make test
# or: go test -count=1 ./...

# Run TypeScript-Go frontend adapter tests
make test-frontend
# or: go test -count=1 ./internal/typescriptgo/...

# Run Node.js Parity Benchmark across all corpus test cases
make test-parity
# or: go run ./cmd/parity

# Build the CLI binary
make build
# or: go build ./cmd/scriptgo

# Run linter
make lint
```

---

## 4. Architecture & Module Responsibilities

ScriptGo enforces strict architectural module boundaries (documented in [`docs/application-structure.md`](docs/application-structure.md)):

```text
cmd/scriptgo -> internal/compiler
internal/compiler -> internal/frontend -> internal/typescriptgo -> TypeScript-Go
internal/compiler -> internal/lowering -> internal/ir
internal/compiler -> internal/interpreter -> internal/ir
internal/compiler -> internal/backend/llvm -> internal/ir
```

- **`cmd/scriptgo`**: CLI flags, process exit codes, and user-facing error reporting.
- **`internal/compiler`**: Pipeline orchestration, artifact selection, invoking Clang/Zig CC.
- **`internal/frontend`**: Program creation, reachable module graph, source diagnostics via TypeScript-Go.
- **`internal/lowering`**: Native subset validation and conversion of AST to typed backend-independent IR.
- **`internal/ir`**: Types, instructions, module structures, and verification invariants.
- **`internal/interpreter`**: Reference tree/instruction execution of verified IR.
- **`internal/backend/llvm`**: Translation of verified IR into LLVM IR and target-specific emission.
- **`internal/runtime`**: Native C runtime library and ABI contract.

---

## 5. Adding Corpus Test Cases

For every new syntax construct, operator, standard library API, or bug fix:

1. Add a test case under `internal/compiler/testdata/corpus/<category>/<feature_name>/`:
   - `main.ts`: The TypeScript test source code.
   - `run.expected`: Expected stdout matching Node.js execution.
   - *(Optional)* `run.err` or `check.err`: Expected error outputs when testing diagnostics.
2. Verify parity against Node.js:
   ```sh
   go run ./cmd/parity
   ```
3. Update [`docs/typescript-parity-report.md`](docs/typescript-parity-report.md) with updated test metrics and feature status.

---

## 6. Pull Request Guidelines

1. **Keep PRs Focused**: One feature or bug fix per PR.
2. **Include Tests**: Ensure new features include corresponding corpus tests.
3. **Verify Full Test Suite**: All tests must pass before submitting:
   ```sh
   make test
   make test-frontend
   make test-parity
   ```
4. **Commit Messages**: Use clear, conventional commit messages (e.g. `feat: add symbol registry support`, `fix: handle negative bounds in string slice`).

---

## 7. Code of Conduct

Please review and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) in all community interactions.
