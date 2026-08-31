#define SQLITE_THREADSAFE 1
#define SQLITE_ENABLE_JSON1 1
#define SQLITE_ENABLE_SESSION 1
#define SQLITE_ENABLE_PREUPDATE_HOOK 1
#define SQLITE_ENABLE_COLUMN_METADATA 1
#define SQLITE_OMIT_DEPRECATED 1

#include "sqlite3.c"
#include "database.c"
#include "statement.c"
#include "results.c"
#include "session.c"
#include "functions.c"
