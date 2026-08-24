/**
 * Probabilistic Data Structures: Bloom Filter & Dynamic BitSet
 * 
 * Demonstrates:
 * - Uint8Array & DataView binary manipulation
 * - Bitwise operations (shifts, masks, bitwise AND/OR/XOR/NOT)
 * - Hash functions (FNV-1a 32-bit & Murmur3-like integer mixing)
 * - Classes, generic interfaces, and math calculations (false positive probability)
 */

export interface BitSetInterface {
    size: number;
    set(index: number): void;
    clear(index: number): void;
    test(index: number): boolean;
    cardinality(): number; // count set bits
    toHexString(): string;
}

export class BitSet implements BitSetInterface {
    public readonly size: number;
    private buffer: Uint8Array;

    constructor(size: number) {
        if (size <= 0) {
            throw new RangeError("BitSet size must be positive");
        }
        this.size = size;
        const byteCount = Math.ceil(size / 8);
        this.buffer = new Uint8Array(byteCount);
    }

    set(index: number): void {
        if (index < 0 || index >= this.size) {
            throw new RangeError(`BitSet index out of range: ${index}`);
        }
        const byteIdx = Math.floor(index / 8);
        const bitIdx = index % 8;
        this.buffer[byteIdx] |= (1 << bitIdx);
    }

    clear(index: number): void {
        if (index < 0 || index >= this.size) {
            throw new RangeError(`BitSet index out of range: ${index}`);
        }
        const byteIdx = Math.floor(index / 8);
        const bitIdx = index % 8;
        this.buffer[byteIdx] &= ~(1 << bitIdx);
    }

    test(index: number): boolean {
        if (index < 0 || index >= this.size) {
            return false;
        }
        const byteIdx = Math.floor(index / 8);
        const bitIdx = index % 8;
        return (this.buffer[byteIdx] & (1 << bitIdx)) !== 0;
    }

    cardinality(): number {
        let count = 0;
        for (let i = 0; i < this.buffer.length; i++) {
            let b = this.buffer[i];
            // Brian Kernighan's algorithm for counting bits in byte
            while (b > 0) {
                b &= (b - 1);
                count++;
            }
        }
        return count;
    }

    toHexString(): string {
        const hexChars: string[] = [];
        for (let i = 0; i < this.buffer.length; i++) {
            const hex = this.buffer[i].toString(16).padStart(2, "0");
            hexChars.push(hex);
        }
        return hexChars.join("");
    }
}

export class Hasher {
    /**
     * 32-bit FNV-1a Hash
     */
    static fnv1a(str: string, seed: number = 0x811c9dc5): number {
        let hash = seed;
        for (let i = 0; i < str.length; i++) {
            const code = str.charCodeAt(i);
            hash ^= code;
            // 32-bit FNV prime multiplication: hash * 16777619
            hash = Math.imul(hash, 0x01000193);
        }
        return hash >>> 0; // convert to unsigned 32-bit integer
    }

    /**
     * Murmur-style avalanche mixer
     */
    static murmurMix(input: number, seed: number): number {
        let h = (input ^ seed) >>> 0;
        h = Math.imul(h ^ (h >>> 16), 0x85ebca6b);
        h = Math.imul(h ^ (h >>> 13), 0xc2b2ae35);
        return (h ^ (h >>> 16)) >>> 0;
    }
}

export class BloomFilter {
    public readonly size: number;
    public readonly numHashes: number;
    private bitSet: BitSet;
    private itemCount: number = 0;

    constructor(expectedElements: number, falsePositiveRate: number = 0.01) {
        if (expectedElements <= 0) {
            throw new RangeError("Expected elements must be greater than 0");
        }
        if (falsePositiveRate <= 0 || falsePositiveRate >= 1) {
            throw new RangeError("False positive rate must be between 0 and 1");
        }

        // Optimal size: m = - (n * ln(p)) / (ln(2)^2)
        const ln2 = Math.log(2);
        const m = Math.ceil(- (expectedElements * Math.log(falsePositiveRate)) / (ln2 * ln2));
        this.size = Math.max(64, m);

        // Optimal hash functions: k = (m / n) * ln(2)
        const k = Math.round((this.size / expectedElements) * ln2);
        this.numHashes = Math.max(1, Math.min(16, k));

        this.bitSet = new BitSet(this.size);
    }

    private getHashes(item: string): number[] {
        const hash1 = Hasher.fnv1a(item, 0x811c9dc5);
        const hash2 = Hasher.murmurMix(hash1, 0x9e3779b9);

        const indices: number[] = [];
        // Kirsch-Mitzenmacher optimization: g_i(x) = (h1(x) + i * h2(x)) mod m
        for (let i = 0; i < this.numHashes; i++) {
            const combined = (hash1 + Math.imul(i, hash2)) >>> 0;
            indices.push(combined % this.size);
        }
        return indices;
    }

    add(item: string): void {
        const hashes = this.getHashes(item);
        for (let i = 0; i < hashes.length; i++) {
            this.bitSet.set(hashes[i]);
        }
        this.itemCount++;
    }

    has(item: string): boolean {
        const hashes = this.getHashes(item);
        for (let i = 0; i < hashes.length; i++) {
            if (!this.bitSet.test(hashes[i])) {
                return false;
            }
        }
        return true;
    }

    getFillRatio(): number {
        return this.bitSet.cardinality() / this.size;
    }

    currentFalsePositiveRate(): number {
        // p = (1 - e^(-k * n / m))^k
        const fill = this.getFillRatio();
        return Math.pow(fill, this.numHashes);
    }

    stats(): string {
        return [
            `BloomFilter Stats:`,
            `  BitSet Size: ${this.size} bits (${Math.ceil(this.size / 8)} bytes)`,
            `  Hash Functions: ${this.numHashes}`,
            `  Items Inserted: ${this.itemCount}`,
            `  Bits Set: ${this.bitSet.cardinality()} / ${this.size} (${(this.getFillRatio() * 100).toFixed(2)}%)`,
            `  Estimated FP Rate: ${(this.currentFalsePositiveRate() * 100).toFixed(4)}%`
        ].join("\n");
    }
}

// ==========================================
// Demonstration
// ==========================================

function main(): void {
    console.log("=== Bloom Filter & BitSet Demonstration ===");

    // 1. Test BitSet
    const bs = new BitSet(32);
    bs.set(0);
    bs.set(7);
    bs.set(8);
    bs.set(15);
    bs.set(31);

    console.log("BitSet test(0):", bs.test(0));
    console.log("BitSet test(1):", bs.test(1));
    console.log("BitSet test(31):", bs.test(31));
    console.log("BitSet cardinality (set bits count):", bs.cardinality());
    console.log("BitSet Hex representation:", bs.toHexString());

    // 2. Test Bloom Filter
    const filter = new BloomFilter(100, 0.05);

    const insertedWords = [
        "apple", "banana", "cherry", "date", "elderberry",
        "fig", "grape", "honeydew", "kiwi", "lemon",
        "mango", "nectarine", "orange", "papaya", "quince",
        "raspberry", "strawberry", "tangerine", "ugli", "vanilla"
    ];

    console.log(`\nInserting ${insertedWords.length} elements into Bloom Filter...`);
    for (let i = 0; i < insertedWords.length; i++) {
        filter.add(insertedWords[i]);
    }

    console.log("\n" + filter.stats());

    // Check membership for inserted words
    let allPresent = true;
    for (let i = 0; i < insertedWords.length; i++) {
        if (!filter.has(insertedWords[i])) {
            allPresent = false;
            console.error(`Error: Inserted word not found: ${insertedWords[i]}`);
        }
    }
    console.log("\nAll inserted elements present (No false negatives):", allPresent);

    // Check test words not inserted to assess false positives
    const nonInsertedWords = [
        "watermelon", "blueberry", "pineapple", "coconut", "avocado",
        "lime", "peach", "plum", "apricot", "blackberry",
        "cranberry", "durian", "guava", "jackfruit", "kumquat"
    ];

    let falsePositiveCount = 0;
    for (let i = 0; i < nonInsertedWords.length; i++) {
        const item = nonInsertedWords[i];
        const isPresent = filter.has(item);
        if (isPresent) {
            console.log(`[False Positive detected] "${item}" reported as present`);
            falsePositiveCount++;
        }
    }

    console.log(`Tested ${nonInsertedWords.length} absent words. False Positives: ${falsePositiveCount}`);
}

main();
