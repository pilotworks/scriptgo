# Compilation Tiers

`scriptgo` uses three explicit compilation tiers. The tier is selected at each
source site after TypeScript-Go has parsed, resolved, and type-checked the
reachable program graph. A program may contain static code and dynamic
dependencies, but a static site never changes meaning because another site is
dynamic.

| Tier            | Selection               | Execution model                                                                                                                                                                                    | Result when unavailable                                                                    |
| --------------- | ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| **Static**      | Default                 | Compile directly to `scriptgo` IR, LLVM, and native code. No JavaScript engine is linked.                                                                                                          | Try Dynamic only when `--dynamic` is explicitly enabled; otherwise report Unsupported.     |
| **Dynamic**     | Opt-in with `--dynamic` | Execute a JavaScript-compatible dynamic island in embedded QuickJS-ng. This includes eligible `.js`/npm package code, `any`, and other values that cannot be represented safely by the static ABI. | Report Unsupported if the feature is not covered by the dynamic runtime contract.          |
| **Unsupported** | No valid implementation | No code is emitted for the rejected site.                                                                                                                                                          | Compile error with stable code, source span/code frame, and a rewrite hint where possible. |

The key invariant is:

```text
default       -> static native code only
--dynamic     -> static native code + explicit QuickJS-ng dynamic islands
unsupported   -> compile error; never silent fallback or guessed semantics
```

## Static

Static is the safe default and the current MVP. A construct is static only when
the compiler can prove its representation, evaluation order, ownership, and
observable JavaScript behavior. Static output must not depend on a hidden
JavaScript engine or on platform-specific libc behavior that is not part of the
runtime ABI.

Examples include typed arithmetic, promoted builtins, compiler-resolved object
layouts, and generic arrays whose element layout is known to the compiler.
TypeScript annotations do not by themselves make a value static: an `any`
value remains dynamic or unsupported until its use is narrowed and proven.

## Dynamic

`--dynamic` is an explicit compatibility mode for code that needs JavaScript's
runtime value model. QuickJS-ng is embedded only in binaries that request this
mode; the default static binary does not link an engine.

Dynamic islands may contain JavaScript npm dependencies, erased TypeScript,
`any`, dynamic property access, function values, prototype-sensitive behavior,
or other operations that require JavaScript semantics. Native code may call into
an island and receive a runtime value, but every boundary must define boxing,
validation, ownership, exceptions, and conversion back to static values.

Dynamic mode is not permission to compile arbitrary code incorrectly. If a
package or operation requires Node APIs that QuickJS-ng does not provide, it
must either use an explicitly implemented Node service adapter or fail as
Unsupported with an actionable diagnostic.

## Unsupported

Unsupported means that neither the static ABI nor the dynamic runtime can
implement the requested semantics for the selected target. Compilation fails
before backend emission for that site. Diagnostics should include:

- a stable error code and the original source span;
- a code frame describing the rejected construct;
- whether `--dynamic` could make the site eligible;
- a rewrite or type-narrowing hint when one is practical.

There is no implicit fallback from Static to Dynamic, no host-language
reinterpretation, and no generated native code for an Unsupported construct.

## Static Error Codes

Static-mode subset diagnostics use the `SG` namespace and a stable four-digit
code. TypeScript diagnostics keep their existing `TSxxxx` codes; `SGxxxx`
identifies a scriptgo representation or tier decision.

| Range    | Group                           | Use for                                                                                    |
| -------- | ------------------------------- | ------------------------------------------------------------------------------------------ |
| `SG1xxx` | Static semantic / type boundary | `any`/`unknown`, union narrowing, generic specialization, function values, structural flow |
| `SG2xxx` | Static lowering / coverage      | Unlowered stdlib, tuple/Date/Map/Set operations, unsupported language lowering             |
| `SG3xxx` | Target capability               | WASI networking, process spawning, signals, native FFI, platform APIs                      |
| `SG4xxx` | Semantic divergence / safety    | Dense-array traps, checked casts, width-copy behavior, runtime hard traps                  |
| `SG9xxx` | Internal compiler / fallback    | Invariant violation, unreachable state, unclassified rejection                             |

| Code     | Meaning                             | Typical examples                                                                       |
| -------- | ----------------------------------- | -------------------------------------------------------------------------------------- |
| `SG1001` | `any`/`unknown` boundary            | `any` without `--dynamic`, unsupported unchecked `unknown`                             |
| `SG1002` | Union narrowing unsupported         | unresolved union operation or missing narrowing proof                                  |
| `SG1003` | Generic specialization unsupported  | type arguments/layout cannot be resolved statically                                    |
| `SG1004` | Unresolved function value           | unpinned generic function, reassigned callable binding, unresolved dynamic call target |
| `SG1005` | Structural flow unsupported         | dynamic property/prototype flow or incompatible record shape                           |
| `SG2001` | Stdlib member not lowered           | declared API exists but has no Static lowering                                         |
| `SG2002` | Tuple operation not lowered         | unsupported tuple method/index operation                                               |
| `SG2003` | Date operation not lowered          | unsupported constructor, getter, parser, or formatter                                  |
| `SG2004` | Map/Set representation limitation   | unsupported key/element type or operation                                              |
| `SG2005` | Language lowering unsupported       | AST/semantic construct with no Static IR lowering                                    |
| `SG3001` | WASI/network capability unavailable | networking or socket operation on target                                               |
| `SG3002` | Process capability unavailable      | process spawning or child-process operation                                            |
| `SG3003` | Signal capability unavailable       | signal registration or delivery                                                        |
| `SG3004` | Native FFI capability unavailable   | unsupported FFI signature or target ABI                                                |
| `SG3005` | Platform API unavailable            | target-specific API without an adapter                                                 |
| `SG4001` | Dense-array safety divergence       | invalid index or hole-dependent operation                                              |
| `SG4002` | Checked-cast failure                | runtime value does not satisfy a requested cast                                        |
| `SG4003` | Copy-on-width-conversion divergence | structural width conversion changes aliasing                                           |
| `SG4004` | Runtime hard trap                   | non-catchable runtime failure                                                          |
| `SG9001` | Internal/fallback rejection         | invariant violation, unreachable state                                                 |

Example:

```text
SG1001: static mode cannot compile value of type `any`; use a concrete type,
narrow the value before this operation, or enable --dynamic
```

Codes are part of the CLI contract. They must remain stable across wording
changes, while the message, code frame, source span, and rewrite hint may grow.

## Temporary Static Fences

For the current Static policy, the entries in the catalog above are temporarily
unsupported. Everything outside these fences is intended to be Static-supportable
and should not be routed to QuickJS-ng merely because it is inconvenient to
lower. A feature leaves the fence only after its representation, JavaScript
semantics, target behavior, and parity tests are defined.

## Static Limitations To Track

The [ScriptC limitations](https://scriptc.dev/limitations) page is a useful
reference for the boundary, but its shipped surface is broader than the
current scriptgo MVP. The following are valid Static features to add over time,
not automatic reasons to embed QuickJS-ng:

| Area                         | Static policy                                                                                                                             |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `null` and `undefined`       | Support with explicit tagged/nullish representations and JS-correct narrowing; reject only operations whose representation is not proven. |
| `unknown`                    | Permit only after an explicitly supported narrowing/check; otherwise `SG1001`.                                                            |
| Unions                       | Permit only with a compiler proof; otherwise `SG1002`.                                                                                    |
| Generics                     | Permit only after specialization; otherwise `SG1003`.                                                                                     |
| Function values              | Permit only for statically pinned signatures; otherwise `SG1004`.                                                                         |
| Structural flow              | Permit only for proven layouts/aliasing; otherwise `SG1005`.                                                                              |
| Tuples                       | Add operations incrementally; missing operations use `SG2002`.                                                                            |
| Standard library             | Add APIs one by one with Node parity tests; missing lowerings use `SG2001`.                                                               |
| `Date`, `Map`, `Set`, regex  | These are Static-support targets; missing operations use `SG2003`/`SG2004` rather than implicit Dynamic execution.                        |
| `==`/`!=`                    | Permit only proven same-type comparisons and documented nullish idioms; unsupported lowering uses `SG2005`.                               |
| `finally`, generators, async | Add explicit IR/state-machine semantics before promotion; use `SG2005` while absent.                                                      |
| Platform APIs                | Gate by target capability and ABI contract; reject unavailable capabilities with `SG3001`–`SG3005`, not a guessed host implementation.    |

ScriptC also documents deliberate Static divergences such as dense arrays,
UTF-8 storage with UTF-16 string methods, reference-counted ownership, and
runtime traps. These are semantic decisions that need their own scriptgo
specification and tests; they are not automatically `Unsupported`.

## Coverage

The compiler should eventually emit a coverage report for each reachable source
site: `static`, `dynamic`, or `unsupported`. This makes npm adoption measurable,
shows where QuickJS-ng is used, and prevents a successful build from hiding a
large amount of dynamic execution.

## Compatibility Reference

ECMAScript and Node.js observable behavior are the compatibility contract. The
Node.js LTS version used for parity fixtures must be recorded, while QuickJS-ng
is an execution component rather than a replacement semantic specification.
See [`stdlib.md`](stdlib.md), [`native-subset.md`](native-subset.md), and the
[TypeScript Handbook](https://www.typescriptlang.org/docs/handbook/intro.html)
for the surrounding language and library policies.
