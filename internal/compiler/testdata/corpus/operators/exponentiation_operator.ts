// @expect: 1024
// @expect: 27
// @expect: 512
// @expect: 4
const base = 2;
const exp = 10;
const result = base ** exp;
console.log(result);

let num = 3;
num **= 3;
console.log(num);

console.log(2 ** 3 ** 2);
console.log(16 ** 0.5);
