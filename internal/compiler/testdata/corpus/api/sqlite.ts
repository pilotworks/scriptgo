import {
    constants,
    backup,
    StatementSync,
    DatabaseSync
} from "node:sqlite";

// @api: sqlite.constants
// @expect: sql_const: true
console.log("sql_const: " + (typeof constants === "object"));

// @api: sqlite.DatabaseSync
// @api: sqlite.sqlite.DatabaseSync
// @api: new sqlite.DatabaseSync
// @expect: sql_db_inst: true
const db = new DatabaseSync(":memory:");
console.log("sql_db_inst: " + (db instanceof DatabaseSync));

// @api: DatabaseSync.isOpen
// @api: sqlite.DatabaseSync.isOpen
// @expect: sql_db_isOpen: true
console.log("sql_db_isOpen: " + db.isOpen);

// @api: DatabaseSync.isTransaction
// @api: sqlite.DatabaseSync.isTransaction
// @expect: sql_db_isTx: false
console.log("sql_db_isTx: " + db.isTransaction);

// @api: DatabaseSync.location
// @api: sqlite.DatabaseSync.location
// @expect: sql_db_loc: true
console.log("sql_db_loc: " + (db.location !== undefined));

// @api: DatabaseSync.exec
// @api: sqlite.DatabaseSync.exec
// @expect: sql_db_exec: true
db.exec("CREATE TABLE t (id INTEGER);");
console.log("sql_db_exec: true");

// @api: DatabaseSync.function
// @api: sqlite.DatabaseSync.function
// @expect: sql_db_fn: true
db.function("my_fn", () => 42);
console.log("sql_db_fn: true");

// @expect: sql_db_fn_val: true
const fnRow = db.prepare("SELECT my_fn() AS val;").get() as { val: number };
console.log("sql_db_fn_val: " + (typeof fnRow.val === "number"));

// @api: sqlite.StatementSync
// @api: sqlite.sqlite.StatementSync
// @api: new sqlite.StatementSync
// @api: DatabaseSync.prepare
// @api: sqlite.DatabaseSync.prepare
// @expect: sql_stmt_inst: true
const stmt = db.prepare("SELECT 1;");
console.log("sql_stmt_inst: " + (stmt instanceof StatementSync));

// @api: StatementSync.all
// @api: sqlite.StatementSync.all
// @expect: sql_stmt_all: 1
console.log("sql_stmt_all: " + stmt.all().length);

// @api: StatementSync.columns
// @api: sqlite.StatementSync.columns
// @expect: sql_stmt_cols: 1
console.log("sql_stmt_cols: " + stmt.columns().length);

// @api: StatementSync.get
// @api: sqlite.StatementSync.get
// @expect: sql_stmt_get: true
console.log("sql_stmt_get: " + (typeof stmt.get() === "object"));

// @api: StatementSync.iterate
// @api: sqlite.StatementSync.iterate
// @expect: sql_stmt_iter: true
console.log("sql_stmt_iter: " + (stmt.iterate() !== null));

// @api: StatementSync.run
// @api: sqlite.StatementSync.run
// @expect: sql_stmt_run: 0
console.log("sql_stmt_run: " + stmt.run().changes);

// @api: StatementSync.setAllowBareNamedParameters
// @api: sqlite.StatementSync.setAllowBareNamedParameters
// @expect: sql_stmt_setBare: true
stmt.setAllowBareNamedParameters(true);
console.log("sql_stmt_setBare: true");

// @api: StatementSync.setAllowUnknownNamedParameters
// @api: sqlite.StatementSync.setAllowUnknownUnknownParameters
// @api: sqlite.StatementSync.setAllowUnknownNamedParameters
// @expect: sql_stmt_setUnk: true
stmt.setAllowUnknownNamedParameters(true);
console.log("sql_stmt_setUnk: true");

// @api: StatementSync.setReturnArrays
// @api: sqlite.StatementSync.setReturnArrays
// @expect: sql_stmt_setArr: true
stmt.setReturnArrays(true);
console.log("sql_stmt_setArr: true");

// @api: StatementSync.setReadBigInts
// @api: sqlite.StatementSync.setReadBigInts
// @expect: sql_stmt_setBigInt: true
stmt.setReadBigInts(true);
console.log("sql_stmt_setBigInt: true");

// @api: sqlite.Session
// @api: sqlite.sqlite.Session
// @api: new sqlite.Session
// @api: DatabaseSync.createSession
// @api: sqlite.DatabaseSync.createSession
// @expect: sql_sess_inst: true
const sess = db.createSession();
console.log("sql_sess_inst: " + (sess !== null));

// @api: Session.changeset
// @api: sqlite.Session.changeset
// @expect: sql_sess_changeset: 0
console.log("sql_sess_changeset: " + sess.changeset().length);

// @api: Session.patchset
// @api: sqlite.Session.patchset
// @expect: sql_sess_patchset: 0
console.log("sql_sess_patchset: " + sess.patchset().length);

// @api: DatabaseSync.applyChangeset
// @api: sqlite.DatabaseSync.applyChangeset
// @expect: sql_db_applyChangeset: true
console.log("sql_db_applyChangeset: " + db.applyChangeset(new Uint8Array(0)));

// @api: DatabaseSync.[Symbol.dispose]
// @api: sqlite.DatabaseSync.[Symbol.dispose]
// @api: DatabaseSync.close
// @api: sqlite.DatabaseSync.close
// @expect: sql_db_close: false
db.close();
console.log("sql_db_close: " + db.isOpen);
