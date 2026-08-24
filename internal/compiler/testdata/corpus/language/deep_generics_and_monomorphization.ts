// @expect: === Deep Generics & Monomorphization Test ===
// @expect: Number Tree Size: 7
// @expect: Number In-Order: 20, 30, 40, 50, 60, 70, 80
// @expect: Number contains 40: true
// @expect: Number contains 99: false
// @expect: 
// @expect: String Tree Size: 7
// @expect: String In-Order: "alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf"
// @expect: String contains 'charlie': true
// @expect: String contains 'zulu': false
// @expect: 
// @expect: Zipped Pairs: Pair(Alice, 25) | Pair(Bob, 30) | Pair(Charlie, 35)
// @expect: Swapped Pairs: Pair(25, Alice) | Pair(30, Bob) | Pair(35, Charlie)
// @expect: 
// @expect: Generic Higher-Order Map/Filter: Sq(2)=4, Sq(4)=16, Sq(6)=36, Sq(8)=64, Sq(10)=100

// Deep Generics, Monomorphization, Constraints, Generic Data Structures & Methods

interface Comparable<T> {
    compareTo(other: T): number;
}

class NumberWrapper implements Comparable<NumberWrapper> {
    constructor(public value: number) {}

    compareTo(other: NumberWrapper): number {
        return this.value - other.value;
    }

    toString(): string {
        return `${this.value}`;
    }
}

class StringWrapper implements Comparable<StringWrapper> {
    constructor(public value: string) {}

    compareTo(other: StringWrapper): number {
        if (this.value < other.value) return -1;
        if (this.value > other.value) return 1;
        return 0;
    }

    toString(): string {
        return `"${this.value}"`;
    }
}

class TreeNode<T extends Comparable<T>> {
    public left: TreeNode<T> | null = null;
    public right: TreeNode<T> | null = null;

    constructor(public value: T) {}
}

class BinarySearchTree<T extends Comparable<T>> {
    private root: TreeNode<T> | null = null;
    private _size: number = 0;

    insert(value: T): void {
        this.root = this.insertRec(this.root, value);
        this._size++;
    }

    private insertRec(node: TreeNode<T> | null, value: T): TreeNode<T> {
        if (node === null) {
            return new TreeNode(value);
        }
        const cmp = value.compareTo(node.value);
        if (cmp < 0) {
            node.left = this.insertRec(node.left, value);
        } else {
            node.right = this.insertRec(node.right, value);
        }
        return node;
    }

    contains(value: T): boolean {
        let curr = this.root;
        while (curr !== null) {
            const cmp = value.compareTo(curr.value);
            if (cmp === 0) return true;
            if (cmp < 0) {
                curr = curr.left;
            } else {
                curr = curr.right;
            }
        }
        return false;
    }

    inOrderTraversal(): T[] {
        const result: T[] = [];
        this.inOrderRec(this.root, result);
        return result;
    }

    private inOrderRec(node: TreeNode<T> | null, result: T[]): void {
        if (node !== null) {
            this.inOrderRec(node.left, result);
            result.push(node.value);
            this.inOrderRec(node.right, result);
        }
    }

    size(): number {
        return this._size;
    }
}

// Generic Pair & Container
class Pair<K, V> {
    constructor(
        public key: K,
        public value: V
    ) {}

    swap(): Pair<V, K> {
        return new Pair(this.value, this.key);
    }

    toString(): string {
        return `Pair(${this.key}, ${this.value})`;
    }
}

// Generic Free Function with Multiple Type Parameters & Array Operations
function zipArrays<A, B>(first: A[], second: B[]): Pair<A, B>[] {
    const len = Math.min(first.length, second.length);
    const result: Pair<A, B>[] = [];
    for (let i = 0; i < len; i++) {
        result.push(new Pair(first[i], second[i]));
    }
    return result;
}

function filterAndMap<T, R>(items: T[], predicate: (item: T) => boolean, transform: (item: T) => R): R[] {
    const out: R[] = [];
    for (let i = 0; i < items.length; i++) {
        if (predicate(items[i])) {
            out.push(transform(items[i]));
        }
    }
    return out;
}

function main(): void {
    console.log("=== Deep Generics & Monomorphization Test ===");

    // 1. Monomorphization with NumberWrapper
    const numTree = new BinarySearchTree<NumberWrapper>();
    const nums = [50, 30, 70, 20, 40, 60, 80];
    for (let i = 0; i < nums.length; i++) {
        numTree.insert(new NumberWrapper(nums[i]));
    }
    console.log("Number Tree Size:", numTree.size());
    console.log("Number In-Order:", numTree.inOrderTraversal().map(n => n.toString()).join(", "));
    console.log("Number contains 40:", numTree.contains(new NumberWrapper(40)));
    console.log("Number contains 99:", numTree.contains(new NumberWrapper(99)));

    // 2. Monomorphization with StringWrapper
    const strTree = new BinarySearchTree<StringWrapper>();
    const words = ["delta", "bravo", "foxtrot", "alpha", "charlie", "echo", "golf"];
    for (let i = 0; i < words.length; i++) {
        strTree.insert(new StringWrapper(words[i]));
    }
    console.log("\nString Tree Size:", strTree.size());
    console.log("String In-Order:", strTree.inOrderTraversal().map(s => s.toString()).join(", "));
    console.log("String contains 'charlie':", strTree.contains(new StringWrapper("charlie")));
    console.log("String contains 'zulu':", strTree.contains(new StringWrapper("zulu")));

    // 3. Multi-type parameter generics
    const names = ["Alice", "Bob", "Charlie"];
    const ages = [25, 30, 35];
    const pairs = zipArrays(names, ages);
    console.log("\nZipped Pairs:", pairs.map(p => p.toString()).join(" | "));

    const swapped = pairs.map(p => p.swap());
    console.log("Swapped Pairs:", swapped.map(p => p.toString()).join(" | "));

    // 4. Higher order generic transformations
    const numbers = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
    const evenSquares = filterAndMap(
        numbers,
        (n) => n % 2 === 0,
        (n) => `Sq(${n})=${n * n}`
    );
    console.log("\nGeneric Higher-Order Map/Filter:", evenSquares.join(", "));
}

main();
