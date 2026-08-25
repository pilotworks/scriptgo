// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 40
class Matrix<T> {
    private grid: T[][] = [];

    constructor(public rows: number, public cols: number, initialValue: T) {
        for (let r = 0; r < rows; r++) {
            const row: T[] = [];
            for (let c = 0; c < cols; c++) {
                row.push(initialValue);
            }
            this.grid.push(row);
        }
    }

    set(r: number, c: number, val: T): void {
        this.grid[r][c] = val;
    }

    get(r: number, c: number): T {
        return this.grid[r][c];
    }
}

const mat = new Matrix<number>(2, 2, 0);
mat.set(0, 0, 10);
mat.set(0, 1, 20);
mat.set(1, 0, 30);
mat.set(1, 1, 40);

console.log(mat.get(0, 0));
console.log(mat.get(0, 1));
console.log(mat.get(1, 0));
console.log(mat.get(1, 1));
