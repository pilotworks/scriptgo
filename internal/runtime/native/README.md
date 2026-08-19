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
