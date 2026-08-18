# Standard Library Compatibility

`scriptgo` aims for a small, explicit standard-library surface familiar to
Node.js users. This is a compatibility target, not a promise of full Node.js or
npm support. TypeScript-Go remains responsible for parsing and type checking;
`scriptgo` owns the native eligibility policy and runtime mapping.

## Compatibility Contract

Each supported API must document four things:

1. the accepted module specifier and TypeScript signature;
2. the observable behavior that matches the selected Node.js LTS reference;
3. intentional native differences, including target and error behavior;
4. a reference/interpreter test and a native executable test.

Compatibility levels are:

| Level               | Meaning                                                                                                           |
| ------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Source-compatible   | Existing Node-style TypeScript imports and calls type-check with the same signatures.                             |
| Behavior-compatible | Supported inputs produce the same observable values, output, side effects, and error class as the Node reference. |
| Target-compatible   | Behavior is stable for the selected scriptgo target; platform-specific differences are explicit.                  |
| Deferred            | The API is documented as planned but must be rejected before lowering.                                            |

An API is not supported merely because a declaration exists. It becomes part of
the native subset only after its runtime representation, ownership, failure
policy, and parity tests are complete.

## Module Names

Use the bare built-in specifier in scriptgo examples:

```ts
import * as path from "path";
```

The canonical scriptgo name is `path`; `node:path` may be accepted later as a
Node.js compatibility alias when both resolve to the same built-in module.
`node_modules`, package
exports, conditional exports, dynamic `require`, and runtime module loading
remain outside the MVP.

## Surface Plan

The status below describes the intended order, not an implementation promise.

| Node.js area                                | Initial scriptgo surface                                               | Status and constraints                                                                                                               |
| ------------------------------------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `console`                                   | `console.log(value)`                                                   | Supported now for one typed argument: `number`, `string`, or `boolean`. Multi-argument formatting, timers, and streams are deferred. |
| `path`                                      | `join`, `resolve`, `dirname`, `basename`, `extname`, `parse`, `format` | First planned pure module. Start with the host POSIX target; `path.posix` and `path.win32` differences must be explicit.             |
| `process`                                   | Read-only `argv`, `env`, `cwd`, `platform`, `arch`                     | Planned. Startup data and environment ownership must be defined by the runtime ABI; mutation and event APIs are deferred.            |
| `os`                                        | `platform`, `arch`, `EOL`, `tmpdir`                                    | Planned target adapter. Results are target-dependent and must not be fabricated by the interpreter.                                  |
| `url`                                       | URL parsing and selected encoding helpers                              | Deferred until string/object representations and error behavior are available.                                                       |
| `util`                                      | Selected formatting and type predicates                                | Deferred until object and variadic-call semantics are available.                                                                     |
| `fs`                                        | Explicit synchronous local-file operations                             | Deferred and opt-in. Encoding, permissions, errors, and sandbox/working-directory policy must be specified before support.           |
| `buffer`                                    | Byte buffers and encoding operations                                   | Deferred until a byte/ownership ABI exists.                                                                                          |
| `events`, `stream`, `http`, `https`, `net`  | No initial surface                                                     | Deferred until objects, callbacks, async scheduling, and shutdown semantics exist.                                                   |
| `crypto`, `child_process`, `worker_threads` | No initial surface                                                     | Deferred for security, portability, and process-model reasons.                                                                       |

The first useful expansion should be pure, synchronous, deterministic modules
such as `path`. Filesystem, process, network, and asynchronous APIs require
explicit target/runtime policies and must not enter the subset by accident.

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

## Non-Goals

This policy does not make `scriptgo` a Node.js runtime. Full npm compatibility,
JavaScript coercion, arbitrary package resolution, dynamic loading, browser
globals, the event loop, and all Node built-ins remain outside the current
scope until their semantics and native costs are separately specified.
