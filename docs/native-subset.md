# scriptgo Native Subset

This document defines the synchronous MVP subset accepted by the native
lowering stage. TypeScript-Go remains the source of truth for parsing and type
checking; this policy decides which checked constructs can be represented by
the current backend-independent IR.

| Area | Supported | Rejected for the MVP |
| --- | --- | --- |
| Files | Closed local `.ts` module graph | npm packages, JavaScript files, dynamic imports |
| Statements | One variable declaration, expression statement, function declaration, static class declaration, `return` | Branches, loops, `try`/`catch`, `throw`, enums |
| Values | `number`, `string`, `boolean` literals and identifiers, dense `number[]` literals, static objects with number/string fields | `null`, `undefined`, dynamic objects, function values |
| Operators | Numeric arithmetic/comparison and string `+` | Operators requiring coercion or unsupported runtime values |
| Calls | User functions, `console.log` with one typed argument, and only explicitly promoted stdlib calls | Methods, constructors, dynamic call targets, unlisted Node APIs |
| Classes | Final static shapes with literal number/string field initializers and zero-argument `new` | Inheritance, methods, constructors with arguments, mutable field assignment |
| Functions | Synchronous functions with typed parameters and returns | Async/generator functions, overloads, closures |

Unsupported constructs are rejected before IR verification or native backend
generation. Diagnostics identify the source file and the unsupported feature.

Standard-library eligibility is defined in [`stdlib.md`](stdlib.md). A Node.js
API is rejected unless its module, signature, runtime representation, and
parity tests are explicitly promoted into this table.
