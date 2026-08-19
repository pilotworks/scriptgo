package runtime

import _ "embed"

// Source contains the native runtime linked into generated executables.
//
//go:embed native/arrays/runtime.c
var arraySource string

//go:embed native/errors/runtime.c
var errorSource string

//go:embed native/objects/runtime.c
var objectSource string

//go:embed native/output/runtime.c
var outputSource string

//go:embed native/fs/runtime.c
var fsSource string

//go:embed native/process/runtime.c
var processSource string

//go:embed native/crypto/runtime.c
var cryptoSource string

//go:embed native/web/runtime.c
var webSource string

//go:embed native/json/runtime.c
var jsonSource string

//go:embed native/numbers/runtime.c
var numberSource string

//go:embed native/strings/runtime.c
var stringSource string

//go:embed native/closures/runtime.c
var closureSource string

//go:embed native/async/runtime.c
var asyncSource string

//go:embed native/os/runtime.c
var osSource string

var Source = []byte(errorSource + "\n" + outputSource + "\n" + arraySource + "\n" + objectSource + "\n" + numberSource + "\n" + stringSource + "\n" + closureSource + "\n" + asyncSource + "\n" + fsSource + "\n" + processSource + "\n" + osSource + "\n" + cryptoSource + "\n" + webSource + "\n" + jsonSource)

