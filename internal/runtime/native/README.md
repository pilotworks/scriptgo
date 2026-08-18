# Native Runtime Implementations

Native runtime source is grouped by value family:

```text
native/
├── arrays/       # Dense number-array allocation, access, bounds, cleanup
├── strings/      # Future owned/length-aware string operations
├── objects/      # Future object headers, shapes, and property operations
├── errors/       # Future runtime diagnostics and failure transport
└── startup/      # Future process initialization and shutdown
```

Only `arrays/` exists today. Add another directory when its ABI contract and
first native operation are ready. Each implementation is linked by
`internal/compiler` as a toolchain input and must not import frontend, IR, or
interpreter packages.
