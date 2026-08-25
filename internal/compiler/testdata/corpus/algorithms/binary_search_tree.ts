// @expect: 20,30,40,50,60,70,80
// @expect: true
// @expect: false
class BSTNode {
    val: number;
    left: BSTNode | null = null;
    right: BSTNode | null = null;

    constructor(val: number) {
        this.val = val;
    }
}

class BST {
    root: BSTNode | null = null;

    insert(val: number): void {
        this.root = this.insertNode(this.root, val);
    }

    private insertNode(node: BSTNode | null, val: number): BSTNode {
        if (!node) return new BSTNode(val);
        if (val < node.val) {
            node.left = this.insertNode(node.left, val);
        } else {
            node.right = this.insertNode(node.right, val);
        }
        return node;
    }

    search(val: number): boolean {
        let curr = this.root;
        while (curr) {
            if (curr.val === val) return true;
            curr = val < curr.val ? curr.left : curr.right;
        }
        return false;
    }

    inorder(): number[] {
        const res: number[] = [];
        const traverse = (n: BSTNode | null) => {
            if (!n) return;
            traverse(n.left);
            res.push(n.val);
            traverse(n.right);
        };
        traverse(this.root);
        return res;
    }
}

const bst = new BST();
bst.insert(50);
bst.insert(30);
bst.insert(70);
bst.insert(20);
bst.insert(40);
bst.insert(60);
bst.insert(80);

console.log(bst.inorder().join(","));
console.log(bst.search(40));
console.log(bst.search(90));
