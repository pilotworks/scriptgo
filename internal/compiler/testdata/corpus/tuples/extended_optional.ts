// @expect: Alice
// @expect: developer
// @expect: Bob
// @expect: unemployed
type Person = [string, string?];

const p1: Person = ["Alice", "developer"];
console.log(p1[0]);
console.log(p1[1]);

const p2: Person = ["Bob"];
console.log(p2[0]);
console.log(p2[1] ?? "unemployed");
