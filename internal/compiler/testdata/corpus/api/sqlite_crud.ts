import { DatabaseSync } from "node:sqlite";

// @expect: sqlite_crud_start: true
console.log("sqlite_crud_start: true");

const db = new DatabaseSync(":memory:");

// @expect: sqlite_is_open: true
console.log("sqlite_is_open: " + db.isOpen);

db.exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, score REAL);");

db.exec("INSERT INTO users (name, score) VALUES ('Alice', 95.5);");
db.exec("INSERT INTO users (name, score) VALUES ('Bob', 82.0);");

const select = db.prepare("SELECT id, name, score FROM users ORDER BY id;");
const rows = select.all();

// @expect: sqlite_rows_count: 2
console.log("sqlite_rows_count: " + rows.length);

const cols = select.columns();
// @expect: sqlite_cols_count: 3
console.log("sqlite_cols_count: " + cols.length);

// @expect: sqlite_col0_name: id
console.log("sqlite_col0_name: " + cols[0].name);

// @expect: sqlite_col1_name: name
console.log("sqlite_col1_name: " + cols[1].name);

const single = db.prepare("SELECT name FROM users WHERE id = 1;").get();
// @expect: sqlite_single_inst: true
console.log("sqlite_single_inst: " + (single !== undefined));

db.close();

// @expect: sqlite_closed: false
console.log("sqlite_closed: " + db.isOpen);
