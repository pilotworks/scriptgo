# Native Runtime ABI

This directory owns the ABI contract used by generated native code. It is not
the reference interpreter; that implementation lives in
[`internal/interpreter`](../interpreter/).

## Current MVP

The LLVM backend currently uses the host C runtime directly:

- `printf` formats `number` values as decimal output;
- `puts` writes strings and boolean text with a trailing newline;
- Clang supplies the platform startup and linker integration.

These calls are the temporary ABI surface for the synchronous MVP. Future
runtime code belongs here when values need managed strings, arrays, objects,
errors, ownership, startup hooks, or a stable versioned ABI.

## Boundary Rules

- The interpreter may model values for testing, but native runtime values must
  have an explicit layout and ownership policy.
- Lowering and LLVM emission may refer to ABI operations, not interpreter
  implementation details.
- Every native runtime operation needs an IR/backend test and a runtime ABI
  contract before it becomes part of the supported subset.
