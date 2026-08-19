package runtime

import _ "embed"

// Source contains the native runtime linked into generated executables.
//
//go:embed native/arrays/runtime.c
var arraySource string

//go:embed native/objects/runtime.c
var objectSource string

//go:embed native/strings/runtime.c
var stringSource string

var Source = []byte(arraySource + "\n" + objectSource + "\n" + stringSource)
