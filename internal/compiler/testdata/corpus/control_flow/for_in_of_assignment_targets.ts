// @expect: of a
// @expect: of b
// @expect: after b
// @expect: in first
// @expect: in second
// @expect: 3
// @expect: 7
// @expect: 1
// for-in/for-of accept identifier, member, and assignment-pattern targets.
let current = "";
for (current of ["a", "b"]) {
  console.log("of", current);
}
console.log("after", current);
const box = { key: "" };
for (box.key in { first: 1, second: 2 }) {
  console.log("in", box.key);
}
let left = 0;
let right = 0;
for ([left, right] of [[1, 2], [3, 4]]) {
  console.log(left + right);
}
let count = 0;
for ([, ] of [[7, 8]]) {
  count += 1;
}
console.log(count);
