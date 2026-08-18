# scriptgo Native Subset

This document defines the synchronous MVP subset accepted by the native
lowering stage. TypeScript-Go remains the source of truth for parsing and type
checking; this policy decides which checked constructs can be represented by
the current backend-independent IR.

| Area | Supported | Rejected for the MVP |
| --- | --- | --- |
| Files | Closed local `.ts` module graph | npm packages, JavaScript files, dynamic imports |
| Statements | One variable declaration, expression statement, function declaration, `return` | Branches, loops, `try`/`catch`, `throw`, classes, enums |
| Values | `number`, `string`, `boolean` literals and identifiers | `null`, `undefined`, arrays, objects, function values |
| Operators | Numeric arithmetic/comparison and string `+` | Operators requiring coercion or unsupported runtime values |
| Calls | User functions and `console.log` with one argument | Methods, constructors, dynamic call targets, other runtime APIs |
| Functions | Synchronous functions with typed parameters and returns | Async/generator functions, overloads, closures |

Unsupported constructs are rejected before IR verification or native backend
generation. Diagnostics identify the source file and the unsupported feature.
