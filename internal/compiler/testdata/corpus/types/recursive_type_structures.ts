// @expect: depth: 3
// @expect: leaves: 3
interface BinaryNode {
    value: number;
    left?: BinaryNode;
    right?: BinaryNode;
}

function getDepth(node?: BinaryNode): number {
    if (!node) return 0;
    return 1 + Math.max(getDepth(node.left), getDepth(node.right));
}

function countLeaves(node?: BinaryNode): number {
    if (!node) return 0;
    if (!node.left && !node.right) return 1;
    return countLeaves(node.left) + countLeaves(node.right);
}

const tree: BinaryNode = {
    value: 10,
    left: {
        value: 5,
        left: { value: 2 },
        right: { value: 7 }
    },
    right: {
        value: 15
    }
};

console.log("depth: " + getDepth(tree));
console.log("leaves: " + countLeaves(tree));
