// @expect: Sum(1, 3): 15
// @expect: Sum(0, 5): 36
// @expect: After update sum(1, 3): 22
// @expect: After update sum(0, 5): 43
class SegmentTree {
    tree: number[];
    n: number;

    constructor(arr: number[]) {
        this.n = arr.length;
        this.tree = [];
        const treeSize = 4 * this.n;
        for (let i = 0; i < treeSize; i++) {
            this.tree.push(0);
        }
        if (this.n > 0) {
            this.build(arr, 0, 0, this.n - 1);
        }
    }

    private build(arr: number[], node: number, start: number, end: number): void {
        if (start === end) {
            this.tree[node] = arr[start];
            return;
        }
        const mid = Math.floor((start + end) / 2);
        const leftNode = 2 * node + 1;
        const rightNode = 2 * node + 2;

        this.build(arr, leftNode, start, mid);
        this.build(arr, rightNode, mid + 1, end);
        this.tree[node] = this.tree[leftNode] + this.tree[rightNode];
    }

    update(idx: number, val: number): void {
        this.updateNode(0, 0, this.n - 1, idx, val);
    }

    private updateNode(node: number, start: number, end: number, idx: number, val: number): void {
        if (start === end) {
            this.tree[node] = val;
            return;
        }
        const mid = Math.floor((start + end) / 2);
        const leftNode = 2 * node + 1;
        const rightNode = 2 * node + 2;

        if (idx <= mid) {
            this.updateNode(leftNode, start, mid, idx, val);
        } else {
            this.updateNode(rightNode, mid + 1, end, idx, val);
        }
        this.tree[node] = this.tree[leftNode] + this.tree[rightNode];
    }

    query(l: number, r: number): number {
        return this.queryNode(0, 0, this.n - 1, l, r);
    }

    private queryNode(node: number, start: number, end: number, l: number, r: number): number {
        if (r < start || end < l) {
            return 0;
        }
        if (l <= start && end <= r) {
            return this.tree[node];
        }
        const mid = Math.floor((start + end) / 2);
        const leftNode = 2 * node + 1;
        const rightNode = 2 * node + 2;

        const leftSum = this.queryNode(leftNode, start, mid, l, r);
        const rightSum = this.queryNode(rightNode, mid + 1, end, l, r);
        return leftSum + rightSum;
    }
}

const data = [1, 3, 5, 7, 9, 11];
const seg = new SegmentTree(data);

console.log("Sum(1, 3): " + seg.query(1, 3)); // 3+5+7 = 15
console.log("Sum(0, 5): " + seg.query(0, 5)); // 36
seg.update(1, 10); // arr[1] is now 10
console.log("After update sum(1, 3): " + seg.query(1, 3)); // 10+5+7 = 22
console.log("After update sum(0, 5): " + seg.query(0, 5)); // 43
