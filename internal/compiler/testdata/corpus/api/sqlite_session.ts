import { DatabaseSync } from "node:sqlite";

// @expect: sqlite_session_start: true
console.log("sqlite_session_start: true");

const sourceDb = new DatabaseSync(":memory:");
const targetDb = new DatabaseSync(":memory:");

sourceDb.exec("CREATE TABLE data(key INTEGER PRIMARY KEY, value TEXT);");
targetDb.exec("CREATE TABLE data(key INTEGER PRIMARY KEY, value TEXT);");

const session = sourceDb.createSession();

sourceDb.exec("INSERT INTO data (key, value) VALUES (1, 'hello');");
sourceDb.exec("INSERT INTO data (key, value) VALUES (2, 'world');");

const changeset = session.changeset();
// @expect: sqlite_changeset_generated: true
console.log("sqlite_changeset_generated: " + (changeset.length > 0));

const applied = targetDb.applyChangeset(changeset);
// @expect: sqlite_changeset_applied: true
console.log("sqlite_changeset_applied: " + applied);

const targetRows = targetDb.prepare("SELECT key, value FROM data ORDER BY key;").all();
// @expect: sqlite_target_count: 2
console.log("sqlite_target_count: " + targetRows.length);

session.close();
sourceDb.close();
targetDb.close();
