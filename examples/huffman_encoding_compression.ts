/**
 * Huffman Encoding & Data Compression Engine
 * 
 * Demonstrates:
 * - Binary Tree Nodes & Huffman Coding algorithm
 * - Min-Priority Queue for Huffman Tree building
 * - Bit-level Stream packing and unpacking into Uint8Array
 * - Frequency Map analysis using Map<string, number>
 * - Compression ratio calculation and roundtrip lossless verification
 */

export class HuffmanNode {
    char: string | null;
    freq: number;
    left: HuffmanNode | null;
    right: HuffmanNode | null;

    constructor(char: string | null, freq: number, left: HuffmanNode | null = null, right: HuffmanNode | null = null) {
        this.char = char;
        this.freq = freq;
        this.left = left;
        this.right = right;
    }

    isLeaf(): boolean {
        return this.left === null && this.right === null;
    }
}

export class HuffmanPriorityQueue {
    private heap: HuffmanNode[] = [];

    push(node: HuffmanNode): void {
        this.heap.push(node);
        this.bubbleUp(this.heap.length - 1);
    }

    pop(): HuffmanNode | undefined {
        if (this.heap.length === 0) return undefined;
        const top = this.heap[0];
        const bottom = this.heap.pop()!;
        if (this.heap.length > 0) {
            this.heap[0] = bottom;
            this.sinkDown(0);
        }
        return top;
    }

    size(): number {
        return this.heap.length;
    }

    private bubbleUp(index: number): void {
        let current = index;
        while (current > 0) {
            const parent = Math.floor((current - 1) / 2);
            if (this.heap[current].freq < this.heap[parent].freq) {
                const tmp = this.heap[current];
                this.heap[current] = this.heap[parent];
                this.heap[parent] = tmp;
                current = parent;
            } else {
                break;
            }
        }
    }

    private sinkDown(index: number): void {
        let current = index;
        const len = this.heap.length;
        while (true) {
            let left = 2 * current + 1;
            let right = 2 * current + 2;
            let smallest = current;

            if (left < len && this.heap[left].freq < this.heap[smallest].freq) {
                smallest = left;
            }
            if (right < len && this.heap[right].freq < this.heap[smallest].freq) {
                smallest = right;
            }

            if (smallest !== current) {
                const tmp = this.heap[current];
                this.heap[current] = this.heap[smallest];
                this.heap[smallest] = tmp;
                current = smallest;
            } else {
                break;
            }
        }
    }
}

export class BitWriter {
    private bytes: number[] = [];
    private currentByte: number = 0;
    private bitPos: number = 0;
    public totalBits: number = 0;

    writeBit(bit: number): void {
        if (bit) {
            this.currentByte |= (1 << (7 - this.bitPos));
        }
        this.bitPos++;
        this.totalBits++;

        if (this.bitPos === 8) {
            this.bytes.push(this.currentByte);
            this.currentByte = 0;
            this.bitPos = 0;
        }
    }

    writeBits(bitString: string): void {
        for (let i = 0; i < bitString.length; i++) {
            this.writeBit(bitString.charAt(i) === "1" ? 1 : 0);
        }
    }

    toUint8Array(): Uint8Array {
        const result = [...this.bytes];
        if (this.bitPos > 0) {
            result.push(this.currentByte);
        }
        return new Uint8Array(result);
    }
}

export class BitReader {
    private buffer: Uint8Array;
    private totalBits: number;
    private currentBitIdx: number = 0;

    constructor(buffer: Uint8Array, totalBits: number) {
        this.buffer = buffer;
        this.totalBits = totalBits;
    }

    hasMoreBits(): boolean {
        return this.currentBitIdx < this.totalBits;
    }

    readBit(): number {
        if (!this.hasMoreBits()) {
            throw new Error("Unexpected end of bit stream");
        }
        const byteIdx = Math.floor(this.currentBitIdx / 8);
        const bitOffset = this.currentBitIdx % 8;
        const bit = (this.buffer[byteIdx] & (1 << (7 - bitOffset))) !== 0 ? 1 : 0;
        this.currentBitIdx++;
        return bit;
    }
}

export class HuffmanCompressor {
    /**
     * Build frequency table from raw text
     */
    static buildFrequencyMap(text: string): Map<string, number> {
        const freqs = new Map<string, number>();
        for (let i = 0; i < text.length; i++) {
            const ch = text.charAt(i);
            freqs.set(ch, (freqs.get(ch) || 0) + 1);
        }
        return freqs;
    }

    /**
     * Build Huffman Tree from frequency table
     */
    static buildTree(freqs: Map<string, number>): HuffmanNode | null {
        if (freqs.size === 0) return null;

        const pq = new HuffmanPriorityQueue();
        for (const [char, freq] of freqs.entries()) {
            pq.push(new HuffmanNode(char, freq));
        }

        while (pq.size() > 1) {
            const left = pq.pop()!;
            const right = pq.pop()!;
            const parent = new HuffmanNode(null, left.freq + right.freq, left, right);
            pq.push(parent);
        }

        return pq.pop()!;
    }

    /**
     * Generate variable-length prefix code mapping (char -> "0101...")
     */
    static generateCodes(root: HuffmanNode | null): Map<string, string> {
        const codeMap = new Map<string, string>();

        function traverse(node: HuffmanNode | null, currentCode: string) {
            if (!node) return;
            if (node.isLeaf()) {
                codeMap.set(node.char!, currentCode.length === 0 ? "0" : currentCode);
                return;
            }
            traverse(node.left, currentCode + "0");
            traverse(node.right, currentCode + "1");
        }

        traverse(root, "");
        return codeMap;
    }

    /**
     * Compress plain text into binary representation
     */
    static compress(text: string): { compressedData: Uint8Array; totalBits: number; root: HuffmanNode | null } {
        if (text.length === 0) {
            return { compressedData: new Uint8Array(0), totalBits: 0, root: null };
        }

        const freqs = this.buildFrequencyMap(text);
        const root = this.buildTree(freqs);
        const codes = this.generateCodes(root);

        const writer = new BitWriter();
        for (let i = 0; i < text.length; i++) {
            const ch = text.charAt(i);
            const code = codes.get(ch)!;
            writer.writeBits(code);
        }

        return {
            compressedData: writer.toUint8Array(),
            totalBits: writer.totalBits,
            root
        };
    }

    /**
     * Decompress binary stream back to original string using Huffman tree
     */
    static decompress(compressedData: Uint8Array, totalBits: number, root: HuffmanNode | null): string {
        if (!root || totalBits === 0) return "";

        const reader = new BitReader(compressedData, totalBits);
        let out = "";

        while (reader.hasMoreBits()) {
            let current: HuffmanNode = root;
            while (!current.isLeaf()) {
                const bit = reader.readBit();
                if (bit === 0) {
                    current = current.left!;
                } else {
                    current = current.right!;
                }
            }
            out += current.char!;
        }

        return out;
    }
}

// ==========================================
// Demonstration
// ==========================================

function main(): void {
    console.log("=== Huffman Data Compression Engine Demo ===");

    const sampleText = "ScriptGo is an Ahead-Of-Time compiler for TypeScript producing high-performance native machine code via LLVM!";
    console.log("Original Text:\n", sampleText);
    console.log(`Original Size: ${sampleText.length} bytes (${sampleText.length * 8} bits)`);

    // Frequency analysis
    const freqs = HuffmanCompressor.buildFrequencyMap(sampleText);
    const root = HuffmanCompressor.buildTree(freqs);
    const codes = HuffmanCompressor.generateCodes(root);

    console.log(`\nUnique Characters: ${freqs.size}`);
    console.log("Huffman Prefix Codes Sample:");
    let sampleCount = 0;
    for (const [char, code] of codes.entries()) {
        const displayChar = char === " " ? "<SPACE>" : char;
        console.log(`  '${displayChar}' (freq: ${freqs.get(char)}): ${code}`);
        sampleCount++;
        if (sampleCount >= 8) break;
    }

    // Compression
    const { compressedData, totalBits } = HuffmanCompressor.compress(sampleText);
    console.log(`\nCompressed Size: ${compressedData.length} bytes (${totalBits} bits)`);
    const compressionRatio = ((1 - (compressedData.length / sampleText.length)) * 100).toFixed(2);
    console.log(`Space Savings: ${compressionRatio}%`);

    // Decompression & Roundtrip Check
    const decompressed = HuffmanCompressor.decompress(compressedData, totalBits, root);
    console.log("\nDecompressed Text:\n", decompressed);
    console.log("Integrity Check Passed (Lossless Roundtrip):", decompressed === sampleText);
}

main();
