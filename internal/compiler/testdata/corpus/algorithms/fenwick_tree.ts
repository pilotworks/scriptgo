// @expect: 15
// @expect: 11
// @expect: 16
// @expect: 33
class FenwickTree {
    private tree: number[];
    private size: number;

    constructor(n: number) {
        this.size = n;
        this.tree = [];
        for (let i = 0; i <= n; i++) {
            this.tree.push(0);
        }
    }

    update(index: number, delta: number): void {
        let i = index;
        while (i <= this.size) {
            this.tree[i] += delta;
            i += i & (-i);
        }
    }

    query(index: number): number {
        let sum = 0;
        let i = index;
        while (i > 0) {
            sum += this.tree[i];
            i -= i & (-i);
        }
        return sum;
    }

    rangeQuery(left: number, right: number): number {
        return this.query(right) - this.query(left - 1);
    }
}

const ft = new FenwickTree(10);
const values = [3, 2, -1, 6, 5, 4, -3, 3, 7, 2];
for (let i = 0; i < values.length; i++) {
    ft.update(i + 1, values[i]);
}

console.log(ft.query(5));
console.log(ft.rangeQuery(3, 7));
ft.update(4, 5);
console.log(ft.rangeQuery(3, 7));
console.log(ft.query(10));
