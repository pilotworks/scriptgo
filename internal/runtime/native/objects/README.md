# Native Object Storage

This runtime service owns only storage primitives for the static object subset.
It allocates a field-counted opaque block, loads and stores primitive slots by
numeric index, releases the block, and reports native failures.

It does not know TypeScript class names, field names, inheritance, constructors,
object shape policy, or property semantics. `internal/lowering` and the LLVM
backend resolve a validated `ir.ObjectShape` to numeric field indexes before
calling these functions.
