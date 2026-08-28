// @expect: 100
// @expect: null
// @expect: 300
// @expect: 100
class CacheNode<K, V> {
    key: K;
    val: V;
    prev: CacheNode<K, V> | null = null;
    next: CacheNode<K, V> | null = null;

    constructor(key: K, val: V) {
        this.key = key;
        this.val = val;
    }
}

class LRUCache<K, V> {
    private capacity: number;
    private map: Map<K, CacheNode<K, V>> = new Map();
    private head: CacheNode<K, V>;
    private tail: CacheNode<K, V>;

    constructor(capacity: number, dummyKey: K, dummyVal: V) {
        this.capacity = capacity;
        this.head = new CacheNode<K, V>(dummyKey, dummyVal);
        this.tail = new CacheNode<K, V>(dummyKey, dummyVal);
        this.head.next = this.tail;
        this.tail.prev = this.head;
    }

    private remove(node: CacheNode<K, V>): void {
        node.prev!.next = node.next;
        node.next!.prev = node.prev;
    }

    private insertToHead(node: CacheNode<K, V>): void {
        node.next = this.head.next;
        node.prev = this.head;
        this.head.next!.prev = node;
        this.head.next = node;
    }

    get(key: K): V | null {
        if (!this.map.has(key)) return null;
        const node = this.map.get(key)!;
        this.remove(node);
        this.insertToHead(node);
        return node.val;
    }

    put(key: K, val: V): void {
        if (this.map.has(key)) {
            const node = this.map.get(key)!;
            node.val = val;
            this.remove(node);
            this.insertToHead(node);
        } else {
            if (this.map.size >= this.capacity) {
                const lru = this.tail.prev!;
                this.remove(lru);
                this.map.delete(lru.key);
            }
            const newNode = new CacheNode<K, V>(key, val);
            this.map.set(key, newNode);
            this.insertToHead(newNode);
        }
    }

    size(): number {
        return this.map.size;
    }
}

const cache = new LRUCache<string, number>(2, "", 0);
cache.put("a", 100);
cache.put("b", 200);
console.log(cache.get("a"));
cache.put("c", 300);
console.log(cache.get("b"));
console.log(cache.get("c"));
console.log(cache.get("a"));
