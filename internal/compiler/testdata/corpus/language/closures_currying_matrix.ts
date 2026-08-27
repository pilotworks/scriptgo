// @expect: [[2,4],[6,8]]
// @expect: 20
function mapMatrix(matrix: number[][], fn: (x: number) => number): number[][] {
    const result: number[][] = [];
    for (let i = 0; i < matrix.length; i++) {
        const row: number[] = [];
        for (let j = 0; j < matrix[i].length; j++) {
            row.push(fn(matrix[i][j]));
        }
        result.push(row);
    }
    return result;
}

const multiplier = (factor: number) => (x: number) => x * factor;
const doubleFn = multiplier(2);

const inputMatrix = [[1, 2], [3, 4]];
const doubled = mapMatrix(inputMatrix, doubleFn);
console.log(JSON.stringify(doubled));

// Accumulator closure across matrix traversal
let totalSum = 0;
const summer = (val: number) => {
    totalSum += val;
    return totalSum;
};
mapMatrix(doubled, summer);
console.log(totalSum);
