import { DatabaseSync } from "node:sqlite";

// @expect: sqlite_params_test: true
console.log("sqlite_params_test: true");

const db = new DatabaseSync(":memory:");

db.exec("CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, price REAL);");

const insert = db.prepare("INSERT INTO items (id, name, price) VALUES (?, ?, ?);");

// Positional binding via run(1, "Laptop", 1200.5)
const res1 = insert.run(1, "Laptop", 1200.5);
// @expect: insert_res1_changes: 1
console.log("insert_res1_changes: " + res1.changes);

// Positional binding via run(2, "Phone", 800.0)
const res2 = insert.run(2, "Phone", 800.0);
// @expect: insert_res2_changes: 1
console.log("insert_res2_changes: " + res2.changes);

// Positional query via get(1)
const selectOne = db.prepare("SELECT name, price FROM items WHERE id = ?;");
const item1 = selectOne.get(1);
// @expect: item1_exists: true
console.log("item1_exists: " + (item1 !== undefined));

// Positional query via all(1000.0)
const selectPrice = db.prepare("SELECT id, name FROM items WHERE price > ? ORDER BY id;");
const expensive = selectPrice.all(1000.0);
// @expect: expensive_count: 1
console.log("expensive_count: " + expensive.length);

// Named parameter binding via run({ id: 3, name: "Tablet", price: 500.0 })
const insertNamed = db.prepare("INSERT INTO items (id, name, price) VALUES (:id, :name, :price);");
const res3 = insertNamed.run({ id: 3, name: "Tablet", price: 500.0 });
// @expect: insert_res3_changes: 1
console.log("insert_res3_changes: " + res3.changes);

const totalItems = db.prepare("SELECT id FROM items;").all();
// @expect: total_items_count: 3
console.log("total_items_count: " + totalItems.length);

db.close();
