# scriptgo Native Subset And Compatibility Modes

This document defines the current synchronous MVP subset accepted by native
lowering. It is the Static tier described in
[`compilation-tiers.md`](compilation-tiers.md), not the complete product
semantic definition. JavaScript and Node.js behavior are the compatibility
target; TypeScript-Go remains the source of truth for parsing, binding, module
resolution, and type checking.

`scriptgo` has three explicit outcomes for every reachable source site:

1. **Static**: compile directly to native code. This is the default and the
   only tier implemented by the current MVP.
2. **Dynamic**: with the planned `--dynamic` mode, execute an eligible dynamic
   island through embedded QuickJS-ng. Dynamic mode is opt-in and is never
   linked implicitly.
3. **Unsupported**: emit a source-anchored compile error when neither static
   lowering nor the dynamic contract can preserve JavaScript semantics.

The native subset table below is therefore a Static eligibility table. A
construct listed as rejected is not automatically Dynamic; it becomes Dynamic
only when the dynamic runtime explicitly supports it.

| Area | Supported | Rejected for the MVP |
| --- | --- | --- |
| Files | Closed local `.ts` module graph | npm/package graphs, `.js` execution, dynamic imports |
| Statements | One variable declaration, expression statement, function declaration, static class declaration, `return` | Branches, loops, `try`/`catch`, `throw`, enums |
| Values | Primitive literals/identifiers, promoted globals, dense primitive arrays, static objects | `null`, `undefined`, dynamic objects, prototypes, function values |
| Operators | Numeric arithmetic/comparison and string `+` where semantics are proven | Operators requiring JavaScript coercion or unsupported runtime values |
| Calls | User functions, `console.log` with one typed argument, and promoted intrinsics `Math.abs`, `Math.ceil`, `Math.floor`, `Math.trunc` | Methods, constructors, dynamic call targets, unlisted Node APIs |
| Classes | Final static shapes with literal number/string field initializers and zero-argument `new` | Inheritance, methods, constructors with arguments, mutable field assignment |
| Functions | Synchronous functions with typed parameters and returns | Async/generator functions, overloads, closures |

Unsupported constructs are rejected before IR verification or native backend
generation. Static subset diagnostics use the `SGxxxx` catalog in
[`compilation-tiers.md`](compilation-tiers.md); diagnostics identify the code,
source file, source span, and unsupported feature. For example, `any` is not a
native layout and must produce `SG1001` in Static mode rather than being
silently treated as a pointer or boxed C value.

The rejected column describes the current MVP, not a permanent language ban.
Features such as `null`/`undefined`, narrowed unions, monomorphized generics,
`Date`, `Map`, `Set`, and additional standard-library APIs can become Static
when their representation, JavaScript behavior, target support, and parity
tests are defined. Features requiring dynamic property/prototype behavior,
unresolved function dispatch, or JavaScript package execution remain Dynamic
or Unsupported according to the selected mode.

## Builtin Semantics

TypeScript-Go remains the source of truth for whether a global or intrinsic is
declared and type-correct. Lowering separately promotes only a small native
surface. `NaN` and `Infinity` are number constants. The promoted `Math.*`
functions use IEEE-754 `number` values and match JavaScript for the supported
one-argument numeric cases; coercion, omitted arguments, extra arguments, and
the rest of the JavaScript `Math` namespace are not part of the native subset.

Standard-library eligibility is defined in [`stdlib.md`](stdlib.md). A Node.js
API is rejected unless its module, signature, runtime representation, and
parity tests are explicitly promoted into this table.

## Compatibility Requirements For Expansion

Every new native feature must preserve JavaScript behavior for:

- left-to-right evaluation and observable side effects;
- `undefined`, `null`, `NaN`, signed zero, numeric overflow, and coercion;
- strict/abstract equality and property lookup rules;
- array holes, mutation, length updates, and prototype-visible behavior;
- thrown values, error names/messages where Node exposes them, and exit status;
- module format, package resolution, caching, and initialization order.

If the native representation cannot preserve one of these rules, the feature
must use a compatible runtime representation or be rejected. A faster
specialized path is valid only behind guards that fall back to JavaScript-
compatible behavior.
