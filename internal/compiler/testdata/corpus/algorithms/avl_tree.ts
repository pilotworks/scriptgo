// @expect: 10 20 25 30 40 50
// @expect: true
class AVLNode {
    key: number;
    height: number;
    left: AVLNode | null;
    right: AVLNode | null;

    constructor(key: number) {
        this.key = key;
        this.height = 1;
        this.left = null;
        this.right = null;
    }
}

class AVLTree {
    root: AVLNode | null = null;

    private getHeight(node: AVLNode | null): number {
        return node ? node.height : 0;
    }

    private getBalance(node: AVLNode | null): number {
        return node ? this.getHeight(node.left) - this.getHeight(node.right) : 0;
    }

    private rightRotate(y: AVLNode): AVLNode {
        const x = y.left!;
        const T2 = x.right;

        x.right = y;
        y.left = T2;

        y.height = Math.max(this.getHeight(y.left), this.getHeight(y.right)) + 1;
        x.height = Math.max(this.getHeight(x.left), this.getHeight(x.right)) + 1;

        return x;
    }

    private leftRotate(x: AVLNode): AVLNode {
        const y = x.right!;
        const T2 = y.left;

        y.left = x;
        x.right = T2;

        x.height = Math.max(this.getHeight(x.left), this.getHeight(x.right)) + 1;
        y.height = Math.max(this.getHeight(y.left), this.getHeight(y.right)) + 1;

        return y;
    }

    insert(key: number): void {
        this.root = this.insertNode(this.root, key);
    }

    private insertNode(node: AVLNode | null, key: number): AVLNode {
        if (!node) {
            return new AVLNode(key);
        }

        if (key < node.key) {
            node.left = this.insertNode(node.left, key);
        } else if (key > node.key) {
            node.right = this.insertNode(node.right, key);
        } else {
            return node;
        }

        node.height = 1 + Math.max(this.getHeight(node.left), this.getHeight(node.right));
        const balance = this.getBalance(node);

        // Left Left Case
        if (balance > 1 && key < node.left!.key) {
            return this.rightRotate(node);
        }

        // Right Right Case
        if (balance < -1 && key > node.right!.key) {
            return this.leftRotate(node);
        }

        // Left Right Case
        if (balance > 1 && key > node.left!.key) {
            node.left = this.leftRotate(node.left!);
            return this.rightRotate(node);
        }

        // Right Left Case
        if (balance < -1 && key < node.right!.key) {
            node.right = this.rightRotate(node.right!);
            return this.leftRotate(node);
        }

        return node;
    }

    inOrder(node: AVLNode | null, acc: number[]): void {
        if (node) {
            this.inOrder(node.left, acc);
            acc.push(node.key);
            this.inOrder(node.right, acc);
        }
    }

    isBalanced(node: AVLNode | null): boolean {
        if (!node) return true;
        const balance = this.getBalance(node);
        if (Math.abs(balance) > 1) return false;
        return this.isBalanced(node.left) && this.isBalanced(node.right);
    }
}

const tree = new AVLTree();
tree.insert(10);
tree.insert(20);
tree.insert(30);
tree.insert(40);
tree.insert(50);
tree.insert(25);

const traversal: number[] = [];
tree.inOrder(tree.root, traversal);
console.log(traversal.join(" "));
console.log(tree.isBalanced(tree.root));
