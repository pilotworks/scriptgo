# Native Runtime Implementations

Native runtime source is grouped by value family:

```text
native/
├── arrays/       # Generic element-layout allocation, access, bounds, cleanup
├── strings/      # String representation and primitive operations
├── objects/      # Object headers, shapes, and property operations
├── errors/       # Future runtime diagnostics and failure transport
└── startup/      # Future process initialization and shutdown
```

The `arrays/`, `strings/`, and `objects/` families exist today. Add another
directory when its ABI contract and first native operation are ready. Each implementation is linked by
`internal/compiler` as a toolchain input and must not import frontend, IR, or
interpreter packages.

## Runtime Benchmarks

The standalone benchmark covers representative array search, string splitting,
and base64 workloads. Build and run it from the repository root using Clang or `zig cc`:

```sh
# Using Clang
clang -O2 \
  internal/runtime/native/bench/main.c \
  internal/runtime/native/errors/runtime.c \
  internal/runtime/native/arrays/runtime.c \
  internal/runtime/native/strings/runtime.c \
  internal/runtime/native/web/runtime.c \
  -o /tmp/scriptgo-runtime-bench && /tmp/scriptgo-runtime-bench

# Using Zig CC
zig cc -O2 \
  internal/runtime/native/bench/main.c \
  internal/runtime/native/errors/runtime.c \
  internal/runtime/native/arrays/runtime.c \
  internal/runtime/native/strings/runtime.c \
  internal/runtime/native/web/runtime.c \
  -o /tmp/scriptgo-runtime-bench && /tmp/scriptgo-runtime-bench
```
