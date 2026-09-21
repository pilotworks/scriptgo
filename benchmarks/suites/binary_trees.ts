// Binary Trees Benchmark: Deep recursive tree creation, traversal, and GC deallocation (depth 14)

class TreeNode {
    item: number;
    left: TreeNode | null;
    right: TreeNode | null;

    constructor(item: number, left: TreeNode | null = null, right: TreeNode | null = null) {
        this.item = item;
        this.left = left;
        this.right = right;
    }

    check(): number {
        if (this.left === null) {
            return this.item;
        }
        const leftVal = this.left.check();
        const rightVal = this.right !== null ? this.right.check() : 0;
        return this.item + leftVal - rightVal;
    }
}

function bottomUpTree(item: number, depth: number): TreeNode {
    if (depth <= 0) {
        return new TreeNode(item, null, null);
    }
    const left = bottomUpTree(2 * item - 1, depth - 1);
    const right = bottomUpTree(2 * item, depth - 1);
    return new TreeNode(item, left, right);
}

function run(): void {
    const maxDepth = 14;
    const stretchDepth = maxDepth + 1;

    const stretchTree = bottomUpTree(0, stretchDepth);
    const stretchCheck = stretchTree.check();

    const longLivedTree = bottomUpTree(0, maxDepth);

    let checkSum = stretchCheck + longLivedTree.check();

    for (let depth = 4; depth <= maxDepth; depth += 2) {
        const iterations = 1 << (maxDepth - depth + 4);
        let depthSum = 0;
        for (let i = 1; i <= iterations; i++) {
            const tree1 = bottomUpTree(i, depth);
            const tree2 = bottomUpTree(-i, depth);
            depthSum += tree1.check() + tree2.check();
        }
        checkSum = (checkSum + depthSum) & 0x7fffffff;
    }

    console.log("Binary trees completed, checksum:", checkSum);
}

run();
