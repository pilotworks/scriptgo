# scriptgo Native Subset And Compatibility Modes

This document defines the current synchronous MVP subset accepted by native
lowering. It is an implementation boundary, not the product semantic
definition. JavaScript and Node.js behavior are the compatibility target;
TypeScript-Go remains the source of truth for parsing, binding, module
resolution, and type checking.

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
generation. Diagnostics identify the source file and the unsupported feature.

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
