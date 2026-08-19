# Standard Library Compatibility

`scriptgo` targets a Node.js-compatible and JavaScript-compatible standard
library surface so useful npm packages can eventually compile without source
rewrites. The current MVP is only a small synchronous subset; unsupported
behavior must be rejected explicitly until its semantics are implemented.
TypeScript-Go remains responsible for parsing, binding, module resolution, and
type checking; `scriptgo` owns native eligibility, runtime mapping, and parity
verification against Node.js.

## Three-Tier Eligibility

The standard library participates in the same three-tier policy as application
code (see [`compilation-tiers.md`](compilation-tiers.md)):

- **Static:** pure TypeScript stdlib or a versioned native primitive whose
  behavior is proven and compiled directly into the executable.
- **Dynamic:** JavaScript/npm library code that requires JavaScript values or
  reflection, executed through embedded QuickJS-ng only with `--dynamic`.
- **Unsupported:** an API with no implemented Static or Dynamic contract;
  compilation fails with a source-anchored diagnostic.

`console` is a language-facing builtin stdlib surface, not a general-purpose
libc escape hatch. `console.log` may lower to a small `scriptgo_print_*` runtime
primitive for Static code; formatting, timers, streams, and other Node console
behavior need their own semantic contract. The C runtime remains below the
core TypeScript stdlib boundary.

An npm package being written in JavaScript does not automatically make the
whole program Dynamic: native-eligible local code remains Static, while the
package import is a Dynamic island when `--dynamic` is enabled. Without that
flag, the import is Unsupported rather than silently interpreted or compiled
with guessed semantics.

## Compatibility Contract

Compatibility priority is ECMAScript behavior first, Node.js observable behavior
second, and native representation/optimization third. Typed IR and the runtime ABI
must model JavaScript values instead of treating TypeScript types as C layouts.
TypeScript annotations are erased at runtime unless a documented runtime check is
required.

Unless a fixture declares another version, parity tests use Node.js 22.x LTS as
the reference runtime. A compatibility claim must identify the ECMAScript and
Node.js version assumptions it relies on.
Each supported API must document four things:

1. the accepted module specifier and TypeScript signature;
2. the observable behavior that matches the selected Node.js LTS reference;
3. intentional native differences, including target and error behavior;
4. a reference/interpreter test and a native executable test.

Compatibility levels are:

| Level               | Meaning                                                                                                           |
| ------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Source-compatible   | Existing Node-style TypeScript and JavaScript package sources can be parsed, resolved, and type-checked. |
| Behavior-compatible | Supported inputs produce the same observable values, output, side effects, coercions, and error class as Node. |
| Target-compatible   | Behavior is stable for the selected scriptgo target; platform-specific differences are explicit. |
| Deferred            | The API is documented as planned but must be rejected before lowering. |

An API is not supported merely because a declaration exists. It becomes part of
the native subset only after its runtime representation, ownership, failure
policy, and parity tests are complete.

## Module Names

Use the bare built-in specifier in scriptgo examples:

```ts
import * as path from "path";
```

The canonical scriptgo name is `path`; `node:path` is a planned compatibility
alias when both resolve to the same built-in module.
The MVP resolves only the current closed graph. The compatibility roadmap must
add `node_modules`, `package.json`, `exports`/`imports`, conditional exports,
CommonJS, ESM, package self-resolution, and the Node module cache before npm
packages can be considered supported.

## Surface Plan

Package loading is a first-class compatibility area. The roadmap must add
node_modules, package.json, exports/imports conditions, CommonJS, ESM, package
self-resolution, cache behavior, and Node-compatible resolution errors before
npm packages can be considered supported.
The status below describes the intended order, not an implementation promise.

| Node.js area                                | Initial scriptgo surface                                               | Status and constraints                                                                                                               |
| ------------------------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `console`                                   | `console.log(value)`                                                   | Supported now for one typed argument: `number`, `string`, or `boolean`. Multi-argument formatting, timers, and streams are deferred. |
| `path`                                      | `join`, `dirname`, `basename`, `extname`                               | Implemented as versioned TypeScript stdlib source for the host POSIX target. `resolve`, `parse`, `format`, and `path.win32` remain deferred. |
| `process`                                   | Read-only `argv`, `env`, `cwd`, `platform`, `arch`                     | Planned. Startup data and environment ownership must be defined by the runtime ABI; mutation and event APIs are deferred.            |
| `os`                                        | `platform`, `arch`, `EOL`, `tmpdir`                                    | Planned target adapter. Results are target-dependent and must not be fabricated by the interpreter.                                  |
| `url`                                       | URL parsing and selected encoding helpers                              | Deferred until string/object representations and error behavior are available.                                                       |
| `util`                                      | Selected formatting and type predicates                                | Deferred until object and variadic-call semantics are available.                                                                     |
| `fs`                                        | Explicit synchronous local-file operations                             | Deferred and opt-in. Encoding, permissions, errors, and sandbox/working-directory policy must be specified before support.           |
| `buffer`                                    | Byte buffers and encoding operations                                   | Deferred until a byte/ownership ABI exists.                                                                                          |
| `events`, `stream`, `http`, `https`, `net`  | No initial surface                                                     | Deferred until objects, callbacks, async scheduling, and shutdown semantics exist.                                                   |
| `crypto`, `child_process`, `worker_threads` | No initial surface                                                     | Deferred for security, portability, and process-model reasons.                                                                       |

The first useful expansion is the pure, synchronous `path` module. Filesystem,
process, network, and asynchronous APIs require explicit target/runtime policies
and must not enter the subset by accident.

## Stdlib Implementation Rule

Implement standard-library functionality in this order:

```text
C primitives -> core TypeScript stdlib -> higher-level TypeScript stdlib
```

Use the following rule when choosing between C and TypeScript:

- Use TypeScript when the logic can be implemented entirely with primitives
  already provided by the language and runtime.
- Use C when the implementation needs OS access, syscalls, an ABI boundary,
  native memory, platform-specific APIs, or a primitive that TypeScript cannot
  yet represent efficiently.
- Prefer higher-level APIs in TypeScript, calling C only through a small native
  boundary.
- Keep C runtime code minimal; do not put business or library logic in the
  runtime when it can be written in TypeScript.
- As the compiler gains features, gradually move implementations from C to
  TypeScript to dogfood the compiler.

The current repository layout is:

```text
internal/
|-- typescriptgo/
|   `-- stdlib/
|       `-- path.ts              # pure TypeScript stdlib
`-- runtime/
    |-- runtime.go               # embeds native sources for linking
    |-- abi/README.md            # ABI contract
    |-- values/README.md         # managed-value policies
    `-- native/
        |-- arrays/runtime.c    # generic array primitives
        |-- strings/runtime.c   # string primitives
        `-- objects/runtime.c   # object primitives
```

Future modules such as `fs.ts`, `os.ts`, and `process.ts` may be TypeScript
wrappers over new native service families, but they are not part of the current
tree.

## Globals And Intrinsics

TypeScript-Go supplies the declarations for standard globals such as `NaN`,
`Infinity`, and `Math`. The native subset promotes only the operations listed
below; lowering maps them to backend-independent intrinsic names and the
interpreter and LLVM backend must implement the same observable behavior.

| Builtin | Native status | Semantic boundary |
| --- | --- | --- |
| `NaN`, `Infinity` | Supported numeric constants | IEEE-754 values; no JavaScript object boxing |
| `Math.abs`, `Math.ceil`, `Math.floor`, `Math.trunc` | Supported, one numeric argument | Matches JavaScript numeric cases; no coercion or omitted/extra arguments |
| Other `Math.*` APIs | Deferred | Must be explicitly promoted with parity tests |

This distinction is intentional: a symbol being present in TypeScript's
standard declarations makes source code type-checkable, but does not make it
native-eligible. The native subset owns eligibility and rejects unpromoted
operations before IR generation.

## Runtime Ownership

Standard-library support crosses the existing boundaries as follows:

```text
TypeScript source
    -> TypeScript-Go module/type checking
    -> native stdlib eligibility in lowering
    -> typed IR runtime operation
    -> LLVM call or inline target operation
    -> internal/runtime/native service
```

- The frontend does not special-case Node behavior.
- Lowering recognizes only an explicit, versioned stdlib manifest and emits
  backend-independent operations.
- The LLVM backend maps those operations to runtime ABI calls or target code;
  it does not decide whether an API is semantically allowed.
- Native implementations belong under
  `internal/runtime/native/<value-or-service-family>/`.
- The reference interpreter implements the same supported behavior as the
  semantic oracle; it must not import the native runtime.

## Node Parity Rules

- Pin parity claims to a documented Node.js LTS reference version.
- Prefer Node's observable behavior over implementation details such as V8
  object layout or libuv internals.
- Make synchronous/asynchronous behavior explicit; synchronous scriptgo APIs
  must not silently emulate Node callbacks or promises.
- Require explicit encodings and units where Node accepts overloaded values.
- Preserve stable error categories and source spans at the compiler boundary;
  runtime failures must include the operation and relevant argument context.
- Keep platform-specific results behind target adapters and test each selected
  target separately.
- Never silently replace an unsupported Node API with a host-specific behavior.

## Testing Strategy

Every promoted API needs:

1. a Node reference fixture for supported and rejected inputs;
2. an interpreter fixture with matching observable output/error behavior;
3. a native executable fixture for the selected host target;
4. a differential test that compares normalized results, diagnostics, and exit
   status.

For nondeterministic APIs such as time, random values, environment, and process
IDs, tests must inject a deterministic provider or compare only the documented
shape and error behavior. Do not snapshot host-specific values as universal
Node parity.

## Current MVP Limitations

These features are deferred from the MVP but remain compatibility goals: full npm
package resolution, JavaScript coercion, arbitrary objects/prototypes, CommonJS/
ESM interop, dynamic loading, browser globals, the event loop, and the complete
Node builtin surface.
