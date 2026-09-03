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

//go:embed native/zlib/runtime.c
var zlibSource string

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

//go:embed native/regex/runtime.c
var regexSource string

//go:embed native/symbol/runtime.c
var symbolSource string

//go:embed native/date/runtime.c
var dateSource string

//go:embed native/typedarray/runtime.c
var typedarraySource string

//go:embed native/typedarray/atomics.c
var atomicsSource string

//go:embed native/timers/runtime.c
var timersSource string

//go:embed native/map/runtime.c
var mapSource string

//go:embed native/set/runtime.c
var setSource string

//go:embed native/encoding/runtime.c
var encodingSource string

//go:embed native/buffer/runtime.c
var bufferSource string

//go:embed native/child_process/runtime.c
var childProcessSource string

//go:embed native/gc/runtime.c
var gcSource string

//go:embed native/weak/runtime.c
var weakSource string

//go:embed native/intl/runtime.c
var intlSource string

//go:embed native/dns/runtime.c
var dnsSource string

//go:embed native/net/runtime.c
var netSource string

//go:embed native/dgram/runtime.c
var dgramSource string

//go:embed native/tls/runtime.c
var tlsSource string

//go:embed native/tls/root_certificates.h
var tlsRootCertificatesHeader string

//go:embed native/sqlite/sqlite_types.h
var sqliteTypesHeader string

//go:embed native/sqlite/sqlite3.c
var sqlite3Source string

//go:embed native/sqlite/database.c
var sqliteDbSource string

//go:embed native/sqlite/statement.c
var sqliteStmtSource string

//go:embed native/sqlite/results.c
var sqliteResultsSource string

//go:embed native/sqlite/session.c
var sqliteSessionSource string

//go:embed native/sqlite/functions.c
var sqliteFunctionsSource string

var sqliteSource = "#define SQLITE_THREADSAFE 1\n#define SQLITE_ENABLE_JSON1 1\n#define SQLITE_ENABLE_SESSION 1\n#define SQLITE_ENABLE_PREUPDATE_HOOK 1\n#define SQLITE_ENABLE_COLUMN_METADATA 1\n#define SQLITE_OMIT_DEPRECATED 1\n" + sqlite3Source + "\n" + sqliteTypesHeader + "\n" + sqliteDbSource + "\n" + sqliteStmtSource + "\n" + sqliteResultsSource + "\n" + sqliteSessionSource + "\n" + sqliteFunctionsSource

var tlsRootCertificatesSource = "#define NODE_WANT_INTERNALS 1\nstatic const char *scriptgo_tls_bundled_root_certificates[] = {\n" + tlsRootCertificatesHeader + "\n};\n#undef NODE_WANT_INTERNALS\n"

var Source = []byte(errorSource + "\n" + outputSource + "\n" + arraySource + "\n" + typedarraySource + "\n" + atomicsSource + "\n" + bufferSource + "\n" + mapSource + "\n" + setSource + "\n" + encodingSource + "\n" + timersSource + "\n" + gcSource + "\n" + weakSource + "\n" + intlSource + "\n" + dnsSource + "\n" + netSource + "\n" + dgramSource + "\n" + tlsRootCertificatesSource + tlsSource + "\n" + objectSource + "\n" + numberSource + "\n" + stringSource + "\n" + closureSource + "\n" + asyncSource + "\n" + fsSource + "\n" + childProcessSource + "\n" + processSource + "\n" + osSource + "\n" + cryptoSource + "\n" + zlibSource + "\n" + webSource + "\n" + jsonSource + "\n" + regexSource + "\n" + symbolSource + "\n" + dateSource + "\n" + sqliteSource)
