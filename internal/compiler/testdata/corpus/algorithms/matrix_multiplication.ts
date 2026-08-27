// @expect: 19 22
// @expect: 43 50
function multiplyMatrices(a: number[][], b: number[][]): number[][] {
    const rowsA = a.length;
    const colsA = a[0].length;
    const colsB = b[0].length;

    const result: number[][] = [];
    for (let i = 0; i < rowsA; i++) {
        const row: number[] = [];
        for (let j = 0; j < colsB; j++) {
            let sum = 0;
            for (let k = 0; k < colsA; k++) {
                sum += a[i][k] * b[k][j];
            }
            row.push(sum);
        }
        result.push(row);
    }
    return result;
}

const m1: number[][] = [
    [1, 2],
    [3, 4]
];

const m2: number[][] = [
    [5, 6],
    [7, 8]
];

const res = multiplyMatrices(m1, m2);
for (const row of res) {
    console.log(row[0] + " " + row[1]);
}
