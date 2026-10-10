// @expect: 8 ADMIN
// @expect: { name: 'ann', [Symbol(id)]: 7, [Symbol(tag)]: 'admin' }
// @expect: true true 1
// @expect: 2 true true Symbol(id),Symbol(tag)
// @expect: false
// @expect: 1 { name: 'ann', [Symbol(id)]: 7 }
// @expect: key name
// @expect: {"name":"ann"} 1
// @expect: hi
// @expect: 5 { label: 'b', [Symbol(k)]: 5 } true
// Symbol-keyed properties are identified by the symbol value: they read and
// write through the symbol, are listed by Object.getOwnPropertySymbols and
// console.log, and are skipped by Object.keys, for..in and JSON.
const id = Symbol("id");
const tag = Symbol("tag");
const user: { name: string; [id]: number; [tag]?: string } = { name: "ann", [id]: 7, [tag]: "admin" };
console.log(user[id] + 1, (user[tag] ?? "").toUpperCase());
console.log(user);
console.log(id in user, "name" in user, Object.keys(user).length);
const syms = Object.getOwnPropertySymbols(user);
console.log(syms.length, syms[0] === id, syms[1] === tag, syms.map((s) => s.toString()).join(","));
const other = Symbol("id");
console.log(other in user);
delete user[tag];
console.log(Object.getOwnPropertySymbols(user).length, user);
for (const key in user) console.log("key", key);
console.log(JSON.stringify(user), Object.entries(user).length);
const greeter = { [id]: () => "hi" };
console.log(greeter[id]());
const k = Symbol("k");
const box: { [k]?: number; label: string } = { label: "b" };
box[k] = 5;
console.log(box[k], box, Object.getOwnPropertySymbols(box)[0] === k);
