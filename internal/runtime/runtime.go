package runtime

import _ "embed"

// Source contains the native runtime linked into generated executables.
//
//go:embed native/arrays/runtime.c
var arraySource string

//go:embed native/objects/runtime.c
var objectSource string

var Source = []byte(arraySource + "\n" + objectSource)
