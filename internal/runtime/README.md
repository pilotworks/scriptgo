# Native Runtime

The runtime is organized by ABI ownership and native representation, not by
compiler pipeline stage:

```text
internal/runtime/
├── README.md                    # This boundary and contribution rules
├── runtime.go                   # Embeds native assets for compiler linking
├── native/errors/runtime.c      # Status diagnostics and failure bridge
├── native/output/runtime.c      # scriptgo_print_* wrappers around libc output
├── abi/
│   └── README.md                # ABI, layout, ownership, and failure contract
├── native/
│   ├── README.md                # Native implementation rules
│   ├── arrays/runtime.c         # Dense number-array operations
│   ├── strings/runtime.c        # String operations
│   └── objects/runtime.c        # Object operations
└── values/
    └── README.md                # Managed-value policies and shared contracts
```

## Ownership

- `abi/` defines the stable compiler/runtime contract. It does not implement
  TypeScript semantics or native operations.
- `native/` contains linked C/runtime implementations grouped by value family.
  Each family owns its layout details, status behavior, and cleanup operations.
- `values/` records cross-family representation and ownership policies. It must
  not become a catch-all helper package.
- `runtime.go` only embeds native sources so `internal/compiler` can create
  temporary toolchain inputs. It must not expose runtime behavior to Go code.

The runtime must remain separate from frontend parsing, lowering policy, and
reference interpretation. New value families follow this order:

1. Define representation, ownership, failure, and compatibility in `abi/`.
2. Add a focused implementation under `native/<family>/`.
3. Add ABI/native integration tests in the owning compiler/runtime boundary.
4. Make lowering depend on the operation only after those tests pass.

Do not add a directory solely as a placeholder. A new directory must contain
its first contract or implementation and a focused test plan.

Node.js-compatible standard-library modules follow the separate policy in
[`docs/stdlib.md`](../../docs/stdlib.md). Their native service code belongs
under `native/` by value or service family; the runtime package does not own
module resolution or TypeScript declarations.
