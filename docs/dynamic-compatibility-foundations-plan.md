# Dynamic Compatibility Foundations Implementation Plan

## Status

- **Roadmap milestone:** Milestone 8, Explicit Dynamic Compatibility Tier
- **Implementation slice:** 8A, analysis, policy, metadata, and reporting
- **Implementation status:** Completed September 6, 2026
- **Runtime engine:** Not included in this slice
- **Package resolution:** Not included in this slice
- **Default behavior:** Static compilation remains unchanged
- **Successor status:** 8B and the closed Dynamic module-graph slice are implemented; registry installation remains separate roadmap work

## Purpose

This plan establishes the compiler contracts required before embedding
QuickJS-ng or resolving npm packages. It turns the documented Static, Dynamic,
and Unsupported policy into explicit compiler data without allowing an
unimplemented Dynamic site to reach IR or a native backend.

The implementation must provide:

1. A deterministic compatibility decision for each reachable source site.
2. A reusable compatibility report owned by `internal/lowering`.
3. Enforcement that preserves the current Static CLI behavior.
4. An opt-in `--dynamic` compiler mode that is recorded in artifacts.
5. Stable diagnostics for Dynamic-eligible sites while the engine is absent.
6. A deterministic compatibility report emitted through the existing CLI.

This work is deliberately separated from QuickJS-ng integration. The result is
a testable boundary that later phases can consume without moving TypeScript
semantics into the compiler orchestrator, runtime, or LLVM backend.

## Goals

- Make tier decisions explicit rather than inferring them from returned errors.
- Preserve TypeScript-Go as the source of truth for parsing, binding, module
  resolution, type checking, and TypeScript semantics.
- Keep compatibility policy and native subset eligibility in
  `internal/lowering`.
- Collect all relevant decisions instead of stopping at the first unsupported
  node during analysis.
- Keep enforcement deterministic and backward compatible in Static mode.
- Add `--dynamic` without silently accepting or incorrectly lowering Dynamic
  code before a JavaScript engine exists.
- Record the selected compatibility mode and report summary in emitted LLVM IR.
- Provide a stable, machine-readable report for future npm compatibility work.
- Keep Static builds independent from QuickJS-ng and any JavaScript engine.

## Non-Goals

- Embedding, building, or linking QuickJS-ng.
- Executing any Dynamic source site.
- Adding Dynamic IR instructions or native/Dynamic call instructions.
- Defining the boxed-value ABI beyond recording requirements for the next phase.
- Resolving `node_modules`, `package.json`, package exports, or package imports.
- Supporting CommonJS or ESM-to-CommonJS interoperability.
- Adding `.js` source compilation.
- Implementing `eval`, `new Function`, `Proxy`, or dynamic import behavior.
- Expanding Node.js standard-library compatibility.
- Changing the current CLI defaults or output for successful Static programs.

## Architectural Constraints

The implementation must preserve the repository dependency direction:

```text
cmd/scriptgo -> compiler
compiler -> frontend -> typescriptgo -> TypeScript-Go
compiler -> lowering -> ir
compiler -> backend/llvm -> ir
```

Module responsibilities for this slice are:

| Module | Responsibility in this plan | Forbidden responsibility |
| --- | --- | --- |
| `cmd/scriptgo` | Parse `--dynamic`, select report output, print errors, choose exit codes | Classify syntax, inspect TypeScript AST data, decide Dynamic eligibility |
| `internal/compiler` | Carry options, orchestrate analysis/lowering/emission, select artifacts | Define language policy, duplicate the compatibility walker |
| `internal/typescriptgo` | Expose normalized syntax kinds and source spans when upstream data is required | Decide whether a construct is Static or Dynamic |
| `internal/frontend` | Build the checked reachable program used by analysis | Define runtime or backend policy |
| `internal/lowering` | Own tier policy, compatibility analysis, diagnostics, and enforcement | Emit LLVM or parse CLI options |
| `internal/ir` | Remain unchanged in this slice | Depend on TypeScript-Go or contain unexecutable Dynamic placeholders |
| `internal/runtime` | Remain unchanged in this slice | Link or expose a placeholder JavaScript engine |
| `internal/backend/llvm` | Render supplied build metadata | Reclassify source sites or inspect TypeScript syntax |

Before implementation, review the pinned TypeScript-Go revision in
`internal/typescriptgo/go.mod` for any syntax kinds that need normalization.
Only add adapter data that is required to identify a source construct and its
span. Do not reproduce upstream parsing or type rules.

## Core Design

Compatibility analysis and compilation enforcement are separate operations:

```text
checked frontend.Program
        |
        v
AnalyzeCompatibility(program, policy)
        |
        +--> CompatibilityReport: all deterministic site decisions
        |
        v
EnforceCompatibility(report, capabilities)
        |
        +--> success: lowering may continue
        |
        +--> diagnostic: stop before IR generation
```

Analysis answers which tier is required by each source site. Enforcement answers
whether the current compiler/runtime capabilities can execute those decisions.
In Milestone 8A, only Static execution is available, even when `--dynamic` is
enabled. Consequently, Dynamic sites can be classified accurately but must
still fail before IR generation.

This separation prevents three undesirable behaviors:

- treating an error string as the compiler's compatibility model;
- marking a feature Unsupported merely because the Dynamic engine is not linked
  yet;
- allowing placeholder IR or LLVM emission for code that cannot execute.

## Terminology

Use different names for source-site tier and compiler invocation mode:

- **Compatibility tier:** `static`, `dynamic`, or `unsupported`; assigned to a
  source site.
- **Compatibility mode:** `static` or `dynamic-enabled`; selected for the whole
  compiler invocation.
- **Runtime capability:** whether a tier has an executable implementation in the
  current build.

A `dynamic-enabled` invocation does not imply that Dynamic execution exists.
During this slice, it means the analyzer may identify Dynamic-eligible sites and
provide the correct diagnostic path, while enforcement still rejects them.

## Proposed Lowering API

Add the compatibility model under `internal/lowering`:

```go
type CompatibilityTier string

const (
	TierStatic      CompatibilityTier = "static"
	TierDynamic     CompatibilityTier = "dynamic"
	TierUnsupported CompatibilityTier = "unsupported"
)

type CompatibilityMode string

const (
	ModeStatic         CompatibilityMode = "static"
	ModeDynamicEnabled CompatibilityMode = "dynamic-enabled"
)

type CompatibilityPolicy struct {
	Mode CompatibilityMode
}

type CompatibilityDecision struct {
	Tier     CompatibilityTier
	FileName string
	Span     typescriptgo.SourceSpan
	Kind     string
	Code     SubsetCode
	Message  string
	Hint     string
}

type CompatibilitySummary struct {
	Static      int
	Dynamic     int
	Unsupported int
}

type CompatibilityReport struct {
	Format    int
	Mode      CompatibilityMode
	Decisions []CompatibilityDecision
	Summary   CompatibilitySummary
}
```

Exact field names may change during implementation, but the following contract
must remain true:

- decisions are structured data, not preformatted CLI strings;
- every non-Static decision has a stable `SGxxxx` code;
- every decision that corresponds to source syntax has a source path and span;
- reports are sorted deterministically before being returned;
- report consumers do not need TypeScript AST APIs;
- CLI formatting is not stored inside the model.

The initial public package functions should be:

```go
func AnalyzeCompatibility(
	program frontend.Program,
	policy CompatibilityPolicy,
) (CompatibilityReport, error)

func EnforceCompatibility(
	report CompatibilityReport,
	capabilities CompatibilityCapabilities,
) error

func ValidateSubset(program frontend.Program) error
```

`ValidateSubset` remains the compatibility wrapper used by existing callers. It
must preserve Static mode and the current first-diagnostic behavior.

`AnalyzeCompatibility` may return an error for analysis failures such as generic
specialization failures or malformed normalized input. Ordinary tier decisions
belong in the report and are not returned as Go errors.

## Site Identity And Ordering

A compatibility site is a normalized statement or expression that participates
in a tier decision. The initial implementation should classify every node
visited by the existing subset validator, including nodes that are Static.

Each site is identified by:

```text
normalized source path + span start + span length + normalized syntax kind
```

Ordering must be stable:

1. Normalized source path.
2. Span start.
3. Span length.
4. Syntax kind.
5. Diagnostic code.

The analyzer must avoid duplicate decisions when the current validation logic
revisits the same node through type and expression checks. Deduplication should
use an internal typed key rather than concatenated strings.

Generated or synthesized nodes without a usable source span should only appear
when they represent a compiler-owned boundary that users can act on. Such sites
must use a stable enclosing source span where possible. An unanchored decision
must not be introduced merely to simplify traversal.

## Initial Classification Policy

The first implementation should preserve the current native subset and apply
the following policy:

| Construct | Static mode classification | Dynamic-enabled classification | Executable in 8A |
| --- | --- | --- | --- |
| Currently lowerable construct | Static | Static | Yes |
| Unconstrained `any` boundary | Unsupported with `SG1001` and Dynamic hint | Dynamic | No |
| Unresolved function/call target | Unsupported with `SG1004` and Dynamic hint | Dynamic when JavaScript call semantics are sufficient | No |
| Dynamic property/prototype flow | Unsupported with `SG1005` and Dynamic hint | Dynamic when no missing host API is required | No |
| `eval` or `new Function` | Unsupported with a stable code and Dynamic hint | Dynamic | No |
| Dynamic `import()` | Unsupported with a stable code and Dynamic hint | Dynamic | No |
| Missing Static lowering that remains a Static target | Unsupported with existing `SG2xxx` code | Unsupported | No |
| Missing target capability | Unsupported with existing `SG3xxx` code | Unsupported unless a documented Dynamic adapter exists | No |
| Internal/unclassified normalized syntax | Unsupported with `SG9001` | Unsupported with `SG9001` | No |

Do not route a construct to Dynamic merely because Static lowering is
inconvenient. Tuples, `Date`, `Map`, `Set`, platform APIs, and other documented
Static targets keep their existing Static fence codes until their semantics are
implemented or the tier policy is explicitly changed.

The classification table must be tested as data. Avoid scattering mode checks
through unrelated lowering functions.

## Diagnostic Policy

Existing Static diagnostic codes and wording are part of the CLI contract.
Refactoring must not change them unless the change is explicitly documented and
tested.

Add a new diagnostic group for Dynamic execution availability:

| Range | Group | Use |
| --- | --- | --- |
| `SG5xxx` | Dynamic compatibility/runtime boundary | Dynamic-eligible site cannot execute because the engine or boundary capability is unavailable |

The first code should be:

| Code | Meaning | Required behavior |
| --- | --- | --- |
| `SG5001` | Dynamic runtime unavailable | Emitted when `--dynamic` classifies a site as Dynamic but the current build has no Dynamic execution engine |

Example Static-mode diagnostic intent:

```text
SG1001: static mode cannot compile value of type `any`; use a concrete type,
narrow the value before this operation, or enable --dynamic
```

Example Dynamic-enabled diagnostic intent for Milestone 8A:

```text
SG5001: this `any` boundary requires Dynamic execution, but the Dynamic runtime
is not available in this build
```

Requirements:

- preserve the original source span and code frame;
- select the first enforcement diagnostic deterministically;
- distinguish Dynamic eligibility from executable Dynamic support;
- never reuse `SG9001` for an expected missing runtime capability;
- keep wording centralized beside the owning compatibility policy.

## File Split Plan

The implementation must split growing files before adding substantial logic.

### `internal/lowering`

`subset.go` is currently approximately 700 lines and should be divided by
responsibility:

```text
internal/lowering/
├── compatibility_model.go       # Tiers, modes, decisions, report, summary
├── compatibility_analyze.go     # Analysis entry point, ordering, deduplication
├── compatibility_enforce.go     # Capability checks and first diagnostic
├── subset_statements.go         # Statement traversal and classification
├── subset_expressions.go        # Expression traversal and classification
├── subset_types.go              # any/unknown/union/type-boundary decisions
├── diagnostic_codes.go          # Stable SG code constants
└── subset.go                    # Small compatibility wrappers if still useful
```

Moving private functions must preserve behavior and should be committed with
tests in the same PR. Do not combine this split with unrelated lowering cleanup.

### `internal/compiler`

Add a focused file instead of growing `compiler.go`:

```text
internal/compiler/
├── compatibility.go             # Program analysis orchestration and report dump
└── options.go                   # BuildOptions mode field and normalization
```

The compiler may translate `BuildOptions.Dynamic` into a lowering policy. It
must not maintain a second feature classification table.

### `cmd/scriptgo`

`main.go` is already over 700 lines. Add shared compatibility flag handling in
a focused file and move related usage text if necessary:

```text
cmd/scriptgo/
├── compatibility_flags.go       # Shared --dynamic flag registration/options
├── usage.go                     # Optional extraction of command usage text
└── main.go                      # Command dispatch and thin handlers
```

The CLI should continue importing only `internal/compiler`.

### `internal/backend/llvm`

`emit.go` is already large. Move deterministic artifact header formatting to:

```text
internal/backend/llvm/metadata.go
```

The metadata formatter consumes supplied values and must not import frontend,
lowering, or TypeScript-Go packages.

## Compiler And CLI Changes

Add the invocation option:

```go
type BuildOptions struct {
	// Existing fields...
	Dynamic bool
}
```

The compiler maps this boolean to a typed lowering mode. A boolean is sufficient
at the CLI/compiler boundary because only two invocation modes exist in this
slice. Site tiers remain typed values owned by lowering.

Add `--dynamic` consistently to:

```text
scriptgo run
scriptgo build
scriptgo check
scriptgo emit
```

Expected behavior:

| Invocation | Source contains only Static sites | Source contains Dynamic site | Source contains Unsupported site |
| --- | --- | --- | --- |
| Default | Continue normally | Existing Static diagnostic | Existing Unsupported diagnostic |
| `--dynamic` | Continue normally and record dynamic-enabled mode | `SG5001` before IR | Existing Unsupported diagnostic |

`check` currently has a Static-only compiler path. Add an options-aware compiler
entry point and retain the old wrapper:

```go
func Check(entryPath string) error
func CheckWithOptions(entryPath string, options BuildOptions) error
```

Do the same for any artifact helper that otherwise bypasses `BuildOptions`, such
as typed IR dumping. All command paths must analyze the same checked program
with the same compatibility policy.

## Artifact Metadata

Add deterministic LLVM comment metadata supplied through `llvm.Options`:

```llvm
; scriptgo.compatibility-mode = "static"
; scriptgo.compatibility-report-format = "1"
; scriptgo.static-sites = "12"
; scriptgo.dynamic-sites = "0"
; scriptgo.unsupported-sites = "0"
```

Requirements:

- default builds emit `static` mode;
- `--dynamic` emits `dynamic-enabled`, even for an all-Static program;
- counts come from the compatibility report used for enforcement;
- metadata order is stable;
- existing compiler, runtime ABI, target, and source hash metadata remain
  unchanged;
- the LLVM backend does not calculate or reinterpret counts;
- absolute paths do not appear in deterministic metadata.

Static executables produced with or without `--dynamic` during Milestone 8A
must link the same runtime sources and contain no QuickJS-ng symbols.

## Compatibility Report Artifact

Extend the existing emit command with:

```text
scriptgo coverage <entry.ts>
scriptgo coverage <entry.ts> --dynamic
scriptgo coverage <entry.ts> --dynamic --format json -o report.json
```

The default output is a deterministic human-readable whole-program summary.
Detailed output uses `--format json` and `encoding/json`; do not build JSON with
string concatenation.

Proposed schema:

```json
{
  "format": 1,
  "mode": "static",
  "summary": {
    "static": 12,
    "dynamic": 0,
    "unsupported": 1
  },
  "sites": [
    {
      "tier": "unsupported",
      "path": "main.ts",
      "start": 13,
      "length": 3,
      "kind": "variable",
      "code": "SG1001",
      "message": "static mode cannot compile value of type `any`",
      "hint": "use a concrete type, narrow the value, or enable --dynamic"
    }
  ]
}
```

Report requirements:

- `format` is an integer schema version starting at `1`;
- paths are normalized relative paths suitable for reproducible output;
- sites are sorted using the ordering contract above;
- all counts exactly match the site list;
- Static sites are included so the report denominator is meaningful;
- messages do not contain ANSI formatting;
- no timestamps, random identifiers, temporary directories, or absolute paths
  are emitted;
- generating a report does not require LLVM or Clang;
- the report may be emitted even when enforcement would reject compilation.

For `coverage`, report generation is the requested artifact.
The command should successfully write the report when frontend checking and
analysis succeed, even if the report contains Dynamic or Unsupported sites.
Normal `check`, `run`, `build`, `typed-ir`, and `llvm-ir` paths continue to
enforce compatibility and return a non-zero exit code when required.

This distinction lets tooling inspect incompatibilities without claiming that
the program is executable.

## Implementation Sequence

### PR 1: Refactor And Analyzer Model

Scope:

1. Split `internal/lowering/subset.go` by responsibility.
2. Add tier, mode, decision, summary, and report types.
3. Convert statement, expression, and type validation into decision collection.
4. Add stable ordering and deduplication.
5. Reimplement `ValidateSubset` through analysis plus Static enforcement.
6. Preserve all existing diagnostic codes, messages, and spans.

Tests:

- analyzer returns Static decisions for a supported program;
- analyzer collects more than one incompatible site;
- report ordering does not depend on map iteration;
- duplicate traversal does not duplicate a site;
- `ValidateSubset` returns the same first error as before;
- lowering still stops before IR for incompatible programs.

Exit criteria:

- no CLI changes;
- no backend changes;
- no runtime changes;
- all existing tests pass without updated goldens unless a previous golden was
  nondeterministic;
- `subset.go` no longer exceeds the repository file-size guideline.

### PR 2: Dynamic-Enabled Mode And Enforcement

Scope:

1. Add `BuildOptions.Dynamic` and normalization plumbing.
2. Add `--dynamic` to all relevant CLI commands.
3. Add Dynamic eligibility classifications for the initial policy table.
4. Add `SG5xxx` catalog documentation and `SG5001`.
5. Add capabilities-aware enforcement.
6. Ensure Dynamic sites never enter lowering in this phase.

Tests:

- all-Static program succeeds in both modes;
- `any` produces `SG1001` by default;
- the same `any` site is classified Dynamic with `--dynamic`;
- normal compilation with `--dynamic` produces `SG5001` for that site;
- a Static-target feature remains Unsupported rather than becoming Dynamic;
- an always-Unsupported construct remains Unsupported in both modes;
- CLI help lists `--dynamic` consistently;
- CLI exit codes remain stable.

Exit criteria:

- the new flag is truthful and cannot produce placeholder executable behavior;
- default CLI behavior is byte-for-byte unchanged where existing tests assert
  exact output;
- no JavaScript engine or new runtime source is linked.

### PR 3: Metadata And Compatibility Artifact

Scope:

1. Pass the compatibility report summary through compiler orchestration.
2. Extract LLVM metadata formatting from `emit.go`.
3. Add compatibility mode and counts to LLVM metadata.
4. Add the `coverage` command with summary output and deterministic JSON through `--format json`.
5. Add compiler and CLI integration tests for report generation.

Tests:

- LLVM output contains the expected mode and counts;
- metadata output is deterministic across repeated builds;
- compatibility JSON is byte-identical across repeated runs;
- report paths are relative and contain no temporary directory;
- report counts match site entries;
- report generation works without Clang;
- incompatible programs can emit reports but cannot emit typed IR, LLVM IR, or
  executables.

Exit criteria:

- the compiler exposes a stable compatibility artifact;
- backend metadata is a pure rendering concern;
- the report can be consumed by future npm package analysis.

### PR 4: Documentation And Milestone Closure

Scope:

1. Update `docs/compilation-tiers.md` with implemented mode behavior, `SG5xxx`,
   and report schema.
2. Update `docs/roadmap.md` to mark only the completed Milestone 8 foundation
   items.
3. Update `docs/typescript-parity-report.md` and its metrics.
4. Document that Dynamic classification exists but execution remains pending.
5. Record the exact follow-up boundary for QuickJS-ng integration.

Exit criteria:

- documentation does not claim that Dynamic code executes;
- roadmap, parity report, CLI help, diagnostics, and tests agree;
- no placeholder npm, CommonJS, Node API, or engine parity is reported.

Documentation may be updated in the implementation PRs rather than a standalone
PR if the project requires parity documentation to remain synchronized at every
commit. The logical scope above still applies.

## Test Corpus Plan

Add corpus fixtures under `internal/compiler/testdata/corpus/` for every new
observable behavior. Suggested layout:

```text
internal/compiler/testdata/corpus/
└── language/
    └── compatibility_tiers/
        ├── static_program/
        │   ├── main.ts
        │   └── run.expected
        ├── any_static_rejected/
        │   ├── main.ts
        │   └── check.err
        ├── any_dynamic_required/
        │   └── main.ts
        ├── dynamic_call_target/
        │   └── main.ts
        ├── dynamic_import/
        │   └── main.ts
        └── unsupported_static_target/
            ├── main.ts
            └── check.err
```

If the current corpus runner cannot pass `--dynamic` or validate compatibility
artifacts, extend its directive model in `internal/compiler/corpus_test.go`
rather than adding an unrelated standalone harness. Keep the directives focused
on compiler behavior, for example:

```text
// @dynamic
// @emit compatibility
```

Only introduce directive syntax after checking existing corpus conventions and
ensuring old fixtures remain valid.

## Test Matrix

| Area | Case | Expected result |
| --- | --- | --- |
| Static regression | Supported primitive program, default mode | Existing IR, LLVM, executable output, and exit status |
| Static regression | Supported program with `--dynamic` | Same behavior, dynamic-enabled metadata |
| Type boundary | `any`, default mode | `SG1001`, source span, Dynamic hint |
| Type boundary | `any`, dynamic-enabled analysis | Site tier is Dynamic |
| Runtime availability | Dynamic site with `--dynamic` compilation | `SG5001` before IR |
| Function values | Unresolved call target | Static rejection or Dynamic classification according to policy |
| Static fence | Missing `Date`/Map/Set/tuple lowering | Existing `SG2xxx`, never automatically Dynamic |
| Target fence | Unsupported WASI capability | Existing `SG3xxx` in both modes unless adapter is documented |
| Internal fallback | Unknown normalized kind | `SG9001`, deterministic span |
| Reporting | Multiple sites across files | Stable ordering and exact counts |
| Reproducibility | Same source in different temporary roots | No absolute path and identical normalized report |
| Backend isolation | Static build | No QuickJS-ng symbol or library |
| CLI | Help and invalid flag use | Stable usage and exit code `2` for usage errors |

## Verification Commands

Run the repository gates sequentially to avoid concurrent TypeScript-Go builds
exhausting memory:

```sh
go test -count=1 ./...
go test -count=1 ./internal/typescriptgo/...
go build ./cmd/scriptgo
```

Also run focused tests during development:

```sh
go test -count=1 ./internal/lowering
go test -count=1 ./internal/compiler
go test -count=1 ./cmd/scriptgo
go test -count=1 ./internal/backend/llvm
```

When LLVM metadata changes, verify both exact metadata presence and complete
artifact determinism. When CLI behavior changes, test stdout, stderr, and exit
status separately.

## Module Responsibility Review

Perform this review before and after each PR:

| Question | Required answer |
| --- | --- |
| Does CLI code inspect TypeScript syntax or types? | No |
| Does compiler orchestration contain a feature classification table? | No |
| Does frontend decide Static versus Dynamic? | No |
| Does lowering emit LLVM text or choose linker inputs? | No |
| Does IR depend on TypeScript-Go? | No |
| Does LLVM inspect frontend or lowering decisions? | No; it renders supplied metadata only |
| Does Static runtime link a JavaScript engine? | No |
| Can a Dynamic or Unsupported site reach verified IR in 8A? | No |
| Are new diagnostics source-anchored and stable? | Yes |
| Are parity documentation and corpus metrics synchronized? | Yes |

## Risks And Mitigations

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Refactoring changes existing first diagnostic | CLI regression | Snapshot existing diagnostic ordering and preserve it in enforcement tests |
| Analyzer double-counts nested or revisited nodes | Misleading coverage | Typed site key, deterministic deduplication, count invariants |
| Every missing Static lowering is labeled Dynamic | Hidden semantic debt | Central classification table and explicit Static-target tests |
| `--dynamic` appears to promise execution | User confusion | `SG5001`, precise help text, documentation stating runtime is absent |
| Report includes temporary absolute paths | Non-reproducible artifacts | Normalize paths relative to the checked program root and test across temp roots |
| Mode checks spread through lowering | Difficult future engine integration | Central policy and enforcement APIs; no ad hoc `if Dynamic` checks |
| LLVM backend gains language policy | Backend coupling | Pass strings/counts through options and keep metadata rendering pure |
| File splits mix with unrelated refactors | Review risk | Move functions mechanically, then add behavior in focused commits |
| TypeScript-Go lacks a required normalized construct | Incorrect classification | Inspect pinned upstream APIs and extend only the adapter model with spans |
| Global lowering diagnostic state leaks between runs | Flaky tests | Keep compatibility reports return-valued and avoid adding new global state |

## Rollback And Compatibility Strategy

Each implementation PR must leave the Static compiler usable and independently
revertible:

- PR 1 is an internal refactor with no CLI or artifact contract change.
- PR 2 adds an optional flag; removing it restores the previous invocation
  surface without changing IR.
- PR 3 adds metadata fields and a new emit mode; existing emit modes remain
  intact.

Do not change existing `SG1xxx` through `SG4xxx` meanings to simplify rollback.
Do not add IR instructions that are unused until QuickJS-ng lands. Do not add
placeholder runtime symbols whose behavior must later be replaced.

## Completion Criteria For Milestone 8A

Milestone 8A is complete only when all of the following are true:

- every source node visited by the subset gate has a deterministic tier decision;
- compatibility analysis returns all decisions rather than failing at the first
  ordinary incompatibility;
- current Static compilation behavior and diagnostics remain compatible;
- `--dynamic` is available on relevant CLI commands and recorded in metadata;
- Dynamic-eligible code fails with `SG5001` until an engine is available;
- Unsupported code remains distinct from Dynamic-eligible code;
- `coverage` produces a deterministic overall summary and `--format json` produces machine-readable detail;
- Dynamic and Unsupported sites cannot reach typed IR or LLVM generation;
- Static artifacts contain no JavaScript engine linkage;
- required corpus fixtures and package tests exist;
- roadmap and TypeScript parity documentation match implemented behavior;
- all mandatory regression tests and the CLI build pass.

## Follow-Up: Milestone 8C

The first real Dynamic island and its closed local module graph are now
implemented. The next vertical slice should extend package installation and
resolution around this graph contract; it should not duplicate module loading
inside the compiler or runtime.

Completed 8B/8C scope:

1. Specify boxed values, validation, ownership, exception transfer, and call
   conventions in the runtime ABI documentation.
2. Add backend-independent IR instructions for entering and calling a Dynamic
   island only after the ABI is approved.
3. Embed QuickJS-ng only for dynamic-enabled builds containing at least one
   executable Dynamic site.
4. Support one local `.js` ES module exporting a pure function whose parameters
   and return value use a deliberately small primitive boundary.
5. Compare Node.js, QuickJS-ng, the compatibility analyzer, and the native
   boundary with differential fixtures.
6. Prove that the same all-Static input produces an artifact without QuickJS-ng.

npm resolution, CommonJS, package caches, Node service adapters, and broad
`any` interoperability should follow only after this boundary is stable.
