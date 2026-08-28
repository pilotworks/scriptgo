// @expect: 7
// @expect: 12
// @expect: 5
// @expect: 0
// @expect: 2
// @expect: 20,20,30,40
// @expect: 5
// @expect: 10
// @expect: 0
// @expect: 2
// @expect: 2
// Prefix and Postfix increment/decrement in expressions

// 1. Variable prefix/postfix combinations
let a = 5;
let b = a++ + ++a; // 5 + 7 = 12, a becomes 7
console.log(a);
console.log(b);

let c = --a - a--; // 6 - 6 = 0, a becomes 5
console.log(a);
console.log(c);

// 2. In array indexing
const arr = [10, 20, 30, 40];
let idx = 0;
arr[idx++] = ++idx * 10;
console.log(idx);
console.log(arr.join(","));

// 3. In loops
let count = 0;
let sum = 0;
while (count++ < 4) {
    sum += count;
}
console.log(count);
console.log(sum);

// 4. Object property prefix/postfix
class Stat {
    hits: number = 0;
}
const s = new Stat();
const oldHits = s.hits++;
const newHits = ++s.hits;
console.log(oldHits);
console.log(newHits);
console.log(s.hits);
