// @expect: 1
// @expect: 3
// @expect: 5
const items = [1, 2, 3, 4, 5, 6];
const [a, , c, , e] = items;

console.log(a);
console.log(c);
console.log(e);
