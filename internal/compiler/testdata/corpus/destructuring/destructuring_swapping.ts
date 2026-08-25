// @expect: 200
// @expect: 100
let a = 100;
let b = 200;

[a, b] = [b, a];

console.log(a);
console.log(b);
