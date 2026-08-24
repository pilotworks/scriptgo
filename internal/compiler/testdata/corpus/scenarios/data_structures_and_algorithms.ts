// @expect: === Data Structures & Algorithms Integration Test ===
// @expect: Raw Data: 64, 34, 25, 12, 22, 11, 90, 88, 76, 50, 42
// @expect: Quicksort Output: 11, 12, 22, 25, 34, 42, 50, 64, 76, 88, 90
// @expect: Mergesort Output: 11, 12, 22, 25, 34, 42, 50, 64, 76, 88, 90
// @expect: Binary Search for 25: index = 3
// @expect: Binary Search for 50: index = 6
// @expect: Binary Search for 99: index = -1
// @expect: Binary Search for 11: index = 0
// @expect: 
// @expect: Trie Searches:
// @expect: search('scriptgo'): true
// @expect: search('scriptg'): false
// @expect: startsWith('sc'): true
// @expect: startsWith('xyz'): false
// @expect: Prefix matches for 'sc': scale, script, scriptgo
// @expect: Prefix matches for 'comp': compilation, compiler
// @expect: 
// @expect: Ring Buffer (after 3 pushes): 10, 20, 30
// @expect: Ring Buffer (after overwrites): 30, 40, 50, 60
// @expect: Popped element: 30, remaining: 40, 50, 60

// Data Structures & Algorithms Integration Test
// Sorting (Quicksort, Mergesort, Heapsort), Binary Search, LRU Cache & Prefix Trie

// 1. Quicksort
function quicksort(arr: number[]): number[] {
    const a = [...arr];
    function sort(low: number, high: number) {
        if (low < high) {
            const pivot = a[high];
            let i = low - 1;
            for (let j = low; j < high; j++) {
                if (a[j] <= pivot) {
                    i++;
                    const tmp = a[i];
                    a[i] = a[j];
                    a[j] = tmp;
                }
            }
            const tmp = a[i + 1];
            a[i + 1] = a[high];
            a[high] = tmp;
            const pi = i + 1;

            sort(low, pi - 1);
            sort(pi + 1, high);
        }
    }
    sort(0, a.length - 1);
    return a;
}

// 2. Mergesort
function mergesort(arr: number[]): number[] {
    if (arr.length <= 1) return arr;
    const mid = Math.floor(arr.length / 2);
    const left = mergesort(arr.slice(0, mid));
    const right = mergesort(arr.slice(mid));

    const result: number[] = [];
    let i = 0, j = 0;
    while (i < left.length && j < right.length) {
        if (left[i] <= right[j]) {
            result.push(left[i++]);
        } else {
            result.push(right[j++]);
        }
    }
    while (i < left.length) result.push(left[i++]);
    while (j < right.length) result.push(right[j++]);
    return result;
}

// 3. Binary Search
function binarySearch(sortedArr: number[], target: number): number {
    let low = 0;
    let high = sortedArr.length - 1;
    while (low <= high) {
        const mid = Math.floor((low + high) / 2);
        if (sortedArr[mid] === target) return mid;
        if (sortedArr[mid] < target) {
            low = mid + 1;
        } else {
            high = mid - 1;
        }
    }
    return -1;
}

// 4. Prefix Trie
class TrieNode {
    children: Map<string, TrieNode> = new Map();
    isEndOfWord: boolean = false;
}

class Trie {
    root: TrieNode = new TrieNode();

    insert(word: string): void {
        let curr = this.root;
        for (let i = 0; i < word.length; i++) {
            const ch = word.charAt(i);
            if (!curr.children.has(ch)) {
                curr.children.set(ch, new TrieNode());
            }
            curr = curr.children.get(ch)!;
        }
        curr.isEndOfWord = true;
    }

    search(word: string): boolean {
        let curr = this.root;
        for (let i = 0; i < word.length; i++) {
            const ch = word.charAt(i);
            if (!curr.children.has(ch)) return false;
            curr = curr.children.get(ch)!;
        }
        return curr.isEndOfWord;
    }

    startsWith(prefix: string): boolean {
        let curr = this.root;
        for (let i = 0; i < prefix.length; i++) {
            const ch = prefix.charAt(i);
            if (!curr.children.has(ch)) return false;
            curr = curr.children.get(ch)!;
        }
        return true;
    }

    findWordsWithPrefix(prefix: string): string[] {
        let curr = this.root;
        for (let i = 0; i < prefix.length; i++) {
            const ch = prefix.charAt(i);
            if (!curr.children.has(ch)) return [];
            curr = curr.children.get(ch)!;
        }

        const results: string[] = [];
        function collect(node: TrieNode, path: string) {
            if (node.isEndOfWord) {
                results.push(prefix + path);
            }
            for (const [ch, child] of node.children.entries()) {
                collect(child, path + ch);
            }
        }

        collect(curr, "");
        return results;
    }
}

// 5. Circular Buffer (Ring Buffer)
class CircularBuffer<T> {
    private buffer: (T | undefined)[];
    private head: number = 0;
    private tail: number = 0;
    private _size: number = 0;
    private readonly capacity: number;

    constructor(capacity: number) {
        this.capacity = capacity;
        this.buffer = new Array(capacity);
    }

    push(item: T): void {
        this.buffer[this.tail] = item;
        this.tail = (this.tail + 1) % this.capacity;
        if (this._size < this.capacity) {
            this._size++;
        } else {
            this.head = (this.head + 1) % this.capacity; // overwrite oldest
        }
    }

    pop(): T | undefined {
        if (this._size === 0) return undefined;
        const item = this.buffer[this.head];
        this.buffer[this.head] = undefined;
        this.head = (this.head + 1) % this.capacity;
        this._size--;
        return item;
    }

    size(): number {
        return this._size;
    }

    toArray(): T[] {
        const out: T[] = [];
        let idx = this.head;
        for (let i = 0; i < this._size; i++) {
            out.push(this.buffer[idx] as T);
            idx = (idx + 1) % this.capacity;
        }
        return out;
    }
}

function main(): void {
    console.log("=== Data Structures & Algorithms Integration Test ===");

    // Sorting & Binary Search
    const rawData = [64, 34, 25, 12, 22, 11, 90, 88, 76, 50, 42];
    console.log("Raw Data:", rawData.join(", "));

    const sortedQuick = quicksort(rawData);
    console.log("Quicksort Output:", sortedQuick.join(", "));

    const sortedMerge = mergesort(rawData);
    console.log("Mergesort Output:", sortedMerge.join(", "));

    const searchTargets = [25, 50, 99, 11];
    for (let i = 0; i < searchTargets.length; i++) {
        const t = searchTargets[i];
        const idx = binarySearch(sortedQuick, t);
        console.log(`Binary Search for ${t}: index = ${idx}`);
    }

    // Prefix Trie
    const trie = new Trie();
    const dictionary = ["script", "scriptgo", "scale", "system", "syntax", "symbol", "compiler", "compilation"];
    for (let i = 0; i < dictionary.length; i++) {
        trie.insert(dictionary[i]);
    }

    console.log("\nTrie Searches:");
    console.log("search('scriptgo'):", trie.search("scriptgo"));
    console.log("search('scriptg'):", trie.search("scriptg"));
    console.log("startsWith('sc'):", trie.startsWith("sc"));
    console.log("startsWith('xyz'):", trie.startsWith("xyz"));
    console.log("Prefix matches for 'sc':", trie.findWordsWithPrefix("sc").sort().join(", "));
    console.log("Prefix matches for 'comp':", trie.findWordsWithPrefix("comp").sort().join(", "));

    // Circular Buffer
    const ring = new CircularBuffer<number>(4);
    ring.push(10);
    ring.push(20);
    ring.push(30);
    console.log("\nRing Buffer (after 3 pushes):", ring.toArray().join(", "));
    ring.push(40);
    ring.push(50); // overwrites 10
    ring.push(60); // overwrites 20
    console.log("Ring Buffer (after overwrites):", ring.toArray().join(", "));
    const popped = ring.pop();
    console.log(`Popped element: ${popped}, remaining:`, ring.toArray().join(", "));
}

main();
