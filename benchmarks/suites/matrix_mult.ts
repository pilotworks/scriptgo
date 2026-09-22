// Matrix Multiplication Benchmark: 256x256 float64 matrices

function createMatrix(size: number, initial: number): number[][] {
    const m: number[][] = [];
    for (let i = 0; i < size; i++) {
        const row: number[] = [];
        for (let j = 0; j < size; j++) {
            row.push(initial + (i * size + j) * 0.001);
        }
        m.push(row);
    }
    return m;
}

function multiply(a: number[][], b: number[][], size: number): number[][] {
    const res: number[][] = [];
    for (let i = 0; i < size; i++) {
        const row: number[] = [];
        for (let j = 0; j < size; j++) {
            let sum = 0.0;
            for (let k = 0; k < size; k++) {
                sum += a[i][k] * b[k][j];
            }
            row.push(sum);
        }
        res.push(row);
    }
    return res;
}

function run(): void {
    const size = 256;
    const a = createMatrix(size, 1.0);
    const b = createMatrix(size, 2.0);

    const c = multiply(a, b, size);

    let checksum = 0.0;
    for (let i = 0; i < size; i++) {
        checksum += c[i][i];
    }
    console.log("Matrix 256x256 multiplied, diagonal checksum:", checksum.toFixed(2));
}

run();
