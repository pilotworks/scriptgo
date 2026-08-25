// @expect: A
// @expect: B
// @expect: C
class TreeNode<T> {
    value: T;
    left: TreeNode<T> | null = null;
    right: TreeNode<T> | null = null;

    constructor(val: T) {
        this.value = val;
    }
}

function inorder<T>(node: TreeNode<T> | null, visit: (val: T) => void): void {
    if (node === null) return;
    inorder(node.left, visit);
    visit(node.value);
    inorder(node.right, visit);
}

const root = new TreeNode<string>("B");
root.left = new TreeNode<string>("A");
root.right = new TreeNode<string>("C");

inorder(root, val => console.log(val));
