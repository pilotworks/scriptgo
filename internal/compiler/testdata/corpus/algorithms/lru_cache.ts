// @expect: 100
// @expect: -1
// @expect: 300
// @expect: 400
// @expect: 500
class LRUNode {
    key: string;
    value: number;
    prev: LRUNode | null = null;
    next: LRUNode | null = null;

    constructor(key: string, value: number) {
        this.key = key;
        this.value = value;
    }
}

class LRUCache {
    capacity: number;
    size: number = 0;
    head: LRUNode | null = null;
    tail: LRUNode | null = null;
    nodes: LRUNode[] = [];

    constructor(capacity: number) {
        this.capacity = capacity;
    }

    private findNode(key: string): LRUNode | null {
        for (let i = 0; i < this.nodes.length; i++) {
            if (this.nodes[i].key === key) {
                return this.nodes[i];
            }
        }
        return null;
    }

    private removeFromArray(key: string): void {
        const nextNodes: LRUNode[] = [];
        for (let i = 0; i < this.nodes.length; i++) {
            if (this.nodes[i].key !== key) {
                nextNodes.push(this.nodes[i]);
            }
        }
        this.nodes = nextNodes;
    }

    private removeNode(node: LRUNode): void {
        if (node.prev !== null) {
            node.prev.next = node.next;
        } else {
            this.head = node.next;
        }

        if (node.next !== null) {
            node.next.prev = node.prev;
        } else {
            this.tail = node.prev;
        }
        node.prev = null;
        node.next = null;
    }

    private addToHead(node: LRUNode): void {
        node.next = this.head;
        node.prev = null;
        if (this.head !== null) {
            this.head.prev = node;
        }
        this.head = node;
        if (this.tail === null) {
            this.tail = node;
        }
    }

    get(key: string): number {
        const node = this.findNode(key);
        if (node === null) {
            return -1;
        }
        this.removeNode(node);
        this.addToHead(node);
        return node.value;
    }

    put(key: string, value: number): void {
        const existing = this.findNode(key);
        if (existing !== null) {
            existing.value = value;
            this.removeNode(existing);
            this.addToHead(existing);
            return;
        }

        if (this.size >= this.capacity) {
            if (this.tail !== null) {
                const tailKey = this.tail.key;
                this.removeNode(this.tail);
                this.removeFromArray(tailKey);
                this.size--;
            }
        }

        const newNode = new LRUNode(key, value);
        this.addToHead(newNode);
        this.nodes.push(newNode);
        this.size++;
    }
}

const cache = new LRUCache(3);
cache.put("a", 100);
cache.put("b", 200);
cache.put("c", 300);

console.log(cache.get("a")); // 100
cache.put("d", 400); // evicts b
console.log(cache.get("b")); // -1
console.log(cache.get("c")); // 300
console.log(cache.get("d")); // 400
cache.put("a", 500); // updates a
console.log(cache.get("a")); // 500
