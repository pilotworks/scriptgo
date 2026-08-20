let x = 10;
let y = ++x;
console.log(x); // 11
console.log(y); // 11

let z = x++;
console.log(x); // 12
console.log(z); // 11

let a = 20;
let b = --a;
console.log(a); // 19
console.log(b); // 19

let c = a--;
console.log(a); // 18
console.log(c); // 19

// In expressions
let m = 5;
let n = 2;
let res = (m++) * (++n) + (++m);
console.log(m); // 7
console.log(n); // 3
console.log(res); // 5 * 3 + 7 = 22
