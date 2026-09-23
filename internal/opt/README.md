# Middle-End Typed IR Optimizer (`internal/opt`)

The `internal/opt` package implements ScriptGo's target-independent Typed IR optimization pipeline. It operates directly on backend-independent [`ir.Module`](../ir/ir.go) values after AST lowering and prior to native backend code emission (LLVM IR or future C generation).

## Architectural Boundaries

Per [`AGENTS.md`](../../AGENTS.md) and [`docs/application-structure.md`](../../docs/application-structure.md):
- **May own**: Target-independent Typed IR optimization passes, dead code elimination, constant folding & algebraic simplification, common subexpression elimination, loop-invariant code motion.
- **Must not own**: TypeScript AST or type checking rules (owned by `internal/frontend` and `internal/typescriptgo`), LLVM IR emission (owned by `internal/backend/llvm`), or runtime ABI/C implementation (owned by `internal/runtime`).

Dependency direction:
```text
cmd/scriptgo -> internal/compiler -> internal/opt -> internal/ir
```

## Invariants and Verification

All passes operate on in-memory `ir.Module` AST/SSA structures.
- **Strict Invariant Guarantee**: After every pass execution, [`m.Verify()`](../ir/verify.go) is asserted. If an optimization pass produces invalid IR (e.g. broken SSA dominance, unlinked blocks, type mismatches), compilation fails immediately with an actionable error.
- **Fixed-Point Iteration**: Passes run iteratively in a fixed-point loop (`maxIters = 5`). If no pass modifies the module during an iteration, the optimizer terminates early.

## Optimization Passes

| Pass | Source File | Description |
| --- | --- | --- |
| **Constant Folding** | [`const_fold.go`](const_fold.go) | Evaluates constant arithmetic/bitwise/comparison expressions at compile time. Applies algebraic identities (e.g. `x + 0`, `x * 1`, `x - 0`, `x * 0`). |
| **CSE (Common Subexpression Elimination)** | [`cse.go`](cse.go) | Scans basic blocks to identify equivalent pure expressions and duplicate loads, eliminating redundant computations by reusing previous SSA values. |
| **LICM (Loop-Invariant Code Motion)** | [`licm.go`](licm.go) | Identifies loops via back-edges, determines loop-invariant pure instructions, and hoists them into preheaders. |
| **Temporary Object Regions** | [`temporary_object_region.go`](temporary_object_region.go) | Brackets compiler-proven non-escaping object builder/visitor sequences so their complete graph bypasses tracing GC and is reclaimed together. |
| **DCE (Dead Code Elimination)** | [`dce.go`](dce.go) | Eliminates unused pure SSA instructions whose results have zero uses, and removes unreachable basic blocks. |

## Optimization Levels

Configured via the `--opt-level` CLI flag (or `BuildOptions.OptLevel`):

| Level | Active Passes | Notes |
| --- | --- | --- |
| `-O0` | *(None)* | Optimizer is bypassed entirely; returns lowered IR as-is. |
| `-O1` | `ConstFold` &rarr; `DCE` | Fast baseline passes suitable for fast compilation. |
| `-O2`, `-O3`, `-Os`, `-Oz`, `-Ofast` | `ConstFold` &rarr; `CSE` &rarr; `LICM` &rarr; `TemporaryObjectRegions` &rarr; `DCE` | Full middle-end optimization pipeline running up to 5 iterations. |
