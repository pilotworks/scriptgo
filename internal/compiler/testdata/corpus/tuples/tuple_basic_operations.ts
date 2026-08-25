// @expect: item
// @expect: 42
// @expect: true
// @expect: 3
// @expect: updated
// @expect: 100
let entry: [string, number, boolean] = ["item", 42, true];

console.log(entry[0]);
console.log(entry[1]);
console.log(entry[2]);
console.log(entry.length);

entry[0] = "updated";
entry[1] = 100;
console.log(entry[0]);
console.log(entry[1]);
