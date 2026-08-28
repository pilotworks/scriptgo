// @expect: 2
// @expect: 1
// @expect: 100
// @expect: 200
// @expect: 1:2
// @expect: 3:4
// @expect: 5:6
// @expect: 42
// @expect: 999
// Destructuring assignment to pre-declared variables

// 1. Array swap
let a = 1;
let b = 2;
[a, b] = [b, a];
console.log(a);
console.log(b);

// 2. Object destructuring assignment (with parentheses)
let x = 0;
let y = 0;
({ x, y } = { x: 100, y: 200 });
console.log(x);
console.log(y);

// 3. Destructuring assignment in loop
let p = 0;
let q = 0;
const pairs: [number, number][] = [[1, 2], [3, 4], [5, 6]];
for (const pair of pairs) {
    [p, q] = pair;
    console.log(p + ":" + q);
}

// 4. Object assignment with rename and defaults
let target1 = 0;
let target2 = 0;
({ x: target1, y: target2 = 999 } = { x: 42 });
console.log(target1);
console.log(target2);
