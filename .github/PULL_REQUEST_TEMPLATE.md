## Description

Please provide a summary of the change and which issue it fixes.

Fixes # (issue)

## Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] ⚡ Performance optimization
- [ ] 🧹 Code refactoring / cleanup

## Parity & Architectural Checklist

- [ ] My code adheres to the package ownership and boundary rules in `docs/application-structure.md`.
- [ ] I have added corpus tests under `internal/compiler/testdata/corpus/` matching Node.js output.
- [ ] I have updated `docs/typescript-parity-report.md` if language features or tests were added/modified.
- [ ] `go test -count=1 ./...` passes without regression.
- [ ] `go test -count=1 ./internal/typescriptgo/...` passes.
- [ ] `go run ./cmd/parity` passes with 100% parity against Node.js.
- [ ] `go build ./cmd/scriptgo` builds cleanly.
