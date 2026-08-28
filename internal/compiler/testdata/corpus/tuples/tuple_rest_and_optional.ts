// @expect: name: Alice, age: 30, active: true
// @expect: name: Bob, age: 0, active: false
// @expect: rest count: 3, all true: false

type UserRow = [string, number?, boolean?];

function formatRow(row: UserRow): string {
  const [name, age = 0, active = false] = row;
  return `name: ${name}, age: ${age}, active: ${active}`;
}

console.log(formatRow(["Alice", 30, true]));
console.log(formatRow(["Bob"]));

type FlagsRow = [string, ...boolean[]];
const flags: FlagsRow = ["config", true, false, true];
const [, ...bools] = flags;
let allTrue = true;
for (const b of bools) {
  if (!b) {
    allTrue = false;
  }
}
console.log(`rest count: ${bools.length}, all true: ${allTrue}`);
