// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 4
type Matrix2x2 = [[number, number], [number, number]];

const mat: Matrix2x2 = [
    [1, 2],
    [3, 4]
];

console.log(mat[0][0]);
console.log(mat[0][1]);
console.log(mat[1][0]);
console.log(mat[1][1]);
