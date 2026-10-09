// @expect: {"2":4,"16":3,"alpha":1,"plain":5} 1 3 4
// @expect: {"1":4,"2":2,"b":1,"a":3} 1,2,b,a
// @expect: 1
// @expect: 2
// @expect: b
// @expect: a
// @expect: zero one 1 a
// Computed keys with a compile-time literal value, numeric keys read by
// name, and property order: integer keys ascending first, then the other
// keys in insertion order (values still evaluate in source order).
const key = "alpha";
const hex = 0x10;
const table = { [key]: 1, [hex]: 3, [2]: 4, plain: 5 };
console.log(JSON.stringify(table), table.alpha, table[16], table[2]);
let step = 0;
const ordered = { b: ++step, 2: ++step, a: ++step, 1: ++step };
console.log(JSON.stringify(ordered), Object.keys(ordered).join());
for (const k in ordered) {
  console.log(k);
}
const names = { 0: "zero", 1: "one" };
const pair: [number, string] = [1, "a"];
console.log(names[0], names[1], pair[0], pair[1]);
