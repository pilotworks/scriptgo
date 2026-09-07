# Managed Values

This directory owns the canonical pre-release boxed-value ABI v1 shared by
`unknown`, exceptions, and Dynamic boundaries. The layout is 24 bytes:
`{uint32_t tag, uint32_t flags, uint64_t payload, uint64_t aux}`. It is
defined in `native/values/scriptgo_value.h` and implemented in
`native/values/runtime.c`.

The old local 16-byte layout is intentionally replaced in place because
scriptgo has not released binary artifacts. Cached or temporary artifacts
created before this change must be rebuilt; they are not compatible with the
current runtime.

Do not byte-copy an owned value. Use the value API's clone, move, and release
operations, and keep engine references within their creating context/thread.
