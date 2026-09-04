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

| Area | Supported in Static Mode | Rejected in Static Mode (`SGxxxx`) |
| --- | --- | --- |
| **Files** | Closed local `.ts` module graph, supported Node.js standard modules (34 core modules via `node:*`), canonical module aliases | npm/package graphs without static types, unbundled `.js` execution |
| **Statements** | Variables (`let`, `const`, `var`), expressions, functions, blocks, `if`/`else`, `switch`/`case` (with fallthrough), `while`, `do..while`, `for`, `for..of`, `for..in`, `for await..of`, Labeled statements (`break label`, `continue label`), `return`, `throw`, `try`/`catch`/`finally`, Explicit Resource Management (`using` & `await using`), Destructuring, Spread/Rest, Enums (numeric, string, const enums), Generators (`yield`, `yield*`), Async Generators | Dynamic imports without static specifiers |
| **Values** | `number` (IEEE-754), `bigint` (64-bit), `string` (UTF-8), `boolean`, `symbol` (with Registry `Symbol.for`, `Symbol.keyFor`), `null`/`undefined`, `unknown` (with boxed tagged representation & control-flow narrowing), Tuples, Monomorphized Generics, Multivariant Unions (`T \| null \| undefined`), TypedArrays & `DataView`, `Map` & `Set` (including all 7 ES2024 Set methods), `WeakMap`, `WeakSet`, `WeakRef`, `FinalizationRegistry`, Dense & Generic Arrays, Object records | Unresolved `any` without type narrowing (`SG1001`), untyped prototype mutation |
| **Operators** | Numeric, bitwise & boolean operators, BigInt arithmetic, strict/abstract equality, relational comparisons, logical operators, unary operators, ternary expressions, string concat/interpolation, optional chaining (`?.`, `fn?.()`), nullish coalescing (`??`), spread/rest, bitwise atomics | Dynamic unproven runtime coercion |
| **Calls** | User functions, arrow functions, lexical closures, currying, higher-order methods (`map`, `filter`, `forEach`, `reduce`, `find`, `some`, `every`), polymorphic methods, `console`, `Math`, WHATWG Web globals (`fetch`, `Headers`, `Request`, `Response`, `URL`, `TextEncoder`, `AbortController`), and supported Node standard library modules | Dynamic call targets with untyped dispatch |
| **Classes** | Constructors, properties, static fields/methods, Class Static Blocks (`static { ... }`), Getters/Setters, Inheritance (`extends`, `super`), Polymorphic VTables, `instanceof`, `new` expressions | Dynamic monkey-patching of class prototypes at runtime |
| **Functions** | Synchronous typed functions, `async`/`await` functions, generator functions (`function*`), async generators, lexical closures with environment captures, default/optional/rest parameters, method overloading | Overload implementations without unified signature |

Unsupported constructs are rejected before IR verification or native backend
generation. Static subset diagnostics use the `SGxxxx` catalog in
[`compilation-tiers.md`](compilation-tiers.md); diagnostics identify the code,
source file, source span, and unsupported feature. For example, `any` is not a
native layout and must produce `SG1001` in Static mode rather than being
silently treated as a pointer or boxed C value.

## Builtin Semantics

TypeScript-Go remains the source of truth for whether a global or intrinsic is
declared and type-correct. Lowering separately promotes only a small native
surface. `NaN` and `Infinity` are number constants. The promoted `Math.*`
functions use IEEE-754 `number` values and match JavaScript for the supported
one-argument numeric cases; coercion, omitted arguments, extra arguments, and
the rest of the JavaScript `Math` namespace are not part of the native subset.

For unboxed native numbers (64-bit IEEE-754 `double`), missing/nullish values
arising from short-circuited optional chaining (`?.`) or uninitialized number fields
are represented natively as `NaN`. Nullish coalescing (`??`) and nullish checks
recognize `NaN` as a missing/nullish state to provide compatible fallbacks.

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
