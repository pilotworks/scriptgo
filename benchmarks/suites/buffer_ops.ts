// Buffer Operations Benchmark: Binary writes, reads, slicing and concatenation
import { Buffer } from "node:buffer";

function run(): void {
    const count = 200000;
    const buf = Buffer.alloc(count * 4);

    // 1. Sequential binary writeInt32LE
    for (let i = 0; i < count; i++) {
        buf.writeInt32LE(i * 3 + 7, i * 4);
    }

    // 2. Sequential binary readInt32LE
    let sum = 0;
    for (let i = 0; i < count; i++) {
        sum += buf.readInt32LE(i * 4);
    }

    // 3. Subarray & concat
    const partA = buf.subarray(0, count * 2);
    const partB = buf.subarray(count * 2, count * 4);
    const combined = Buffer.concat([partA, partB]);

    console.log("Buffer ops completed, bytes:", combined.length, "checksum:", sum);
}

run();
