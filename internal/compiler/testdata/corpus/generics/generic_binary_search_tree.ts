// @expect: Search banana: 20
// @expect: Search grape: null
// @expect: In-order traversal:
// @expect:   apple -> 10
// @expect:   banana -> 20
// @expect:   cherry -> 30
// @expect:   date -> 40
type CompareFn<K> = (a: K, b: K) => number;

class BSTNode<K, V> {
    key: K;
    val: V;
    left: BSTNode<K, V> | null = null;
    right: BSTNode<K, V> | null = null;

    constructor(key: K, val: V) {
        this.key = key;
        this.val = val;
    }
}

class GenericBST<K, V> {
    root: BSTNode<K, V> | null = null;
    private compare: CompareFn<K>;

    constructor(compare: CompareFn<K>) {
        this.compare = compare;
    }

    insert(key: K, val: V): void {
        this.root = this.insertNode(this.root, key, val);
    }

    private insertNode(node: BSTNode<K, V> | null, key: K, val: V): BSTNode<K, V> {
        if (node === null) {
            return new BSTNode(key, val);
        }
        const cmp = this.compare(key, node.key);
        if (cmp < 0) {
            node.left = this.insertNode(node.left, key, val);
        } else if (cmp > 0) {
            node.right = this.insertNode(node.right, key, val);
        } else {
            node.val = val;
        }
        return node;
    }

    search(key: K): V | null {
        let curr = this.root;
        while (curr !== null) {
            const cmp = this.compare(key, curr.key);
            if (cmp === 0) {
                return curr.val;
            } else if (cmp < 0) {
                curr = curr.left;
            } else {
                curr = curr.right;
            }
        }
        return null;
    }

    inOrderWalk(visitor: (key: K, val: V) => void): void {
        this.walk(this.root, visitor);
    }

    private walk(node: BSTNode<K, V> | null, visitor: (key: K, val: V) => void): void {
        if (node !== null) {
            this.walk(node.left, visitor);
            visitor(node.key, node.val);
            this.walk(node.right, visitor);
        }
    }
}

const tree = new GenericBST<string, number>((a, b) => {
    if (a < b) return -1;
    if (a > b) return 1;
    return 0;
});

tree.insert("cherry", 30);
tree.insert("apple", 10);
tree.insert("banana", 20);
tree.insert("date", 40);

console.log("Search banana: " + tree.search("banana"));
console.log("Search grape: " + tree.search("grape"));

console.log("In-order traversal:");
tree.inOrderWalk((k, v) => {
    console.log("  " + k + " -> " + v);
});
