// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 4
const matrix = [
    [1, 2],
    [3, 4]
];

const [[a, b], [c, d]] = matrix;

console.log(a);
console.log(b);
console.log(c);
console.log(d);
