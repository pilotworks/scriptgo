// @expect: -42
// @expect: 42
// @expect: 123
// @expect: false
// @expect: true
// @expect: true
// @expect: undefined
const x = 42;
const str = "123";
const bTrue = true;
const bFalse = false;

console.log(-x);
console.log(+x);
console.log(+str);
console.log(!bTrue);
console.log(!bFalse);
console.log(!!x);
console.log(void x);
