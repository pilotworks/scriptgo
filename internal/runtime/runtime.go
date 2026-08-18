package runtime

import _ "embed"

// Source contains the native runtime linked into generated executables.
//
//go:embed native/arrays/runtime.c
var Source []byte
