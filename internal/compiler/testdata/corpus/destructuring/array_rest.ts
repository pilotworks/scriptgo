// @expect: 10
// @expect: 20
// @expect: 3
// @expect: 30
// @expect: 40
// @expect: 50
const numbers = [10, 20, 30, 40, 50];
const [first, second, ...rest] = numbers;

console.log(first);
console.log(second);
console.log(rest.length);
console.log(rest[0]);
console.log(rest[1]);
console.log(rest[2]);
