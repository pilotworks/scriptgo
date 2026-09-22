// Base64 Transcoding Benchmark: 10MB binary encoding & decoding with integrity verification

function run(): void {
    const size = 10 * 1024 * 1024; // 10MB
    const buffer = Buffer.alloc(size);

    // Fill with deterministic pseudo-random bytes
    let seed = 42;
    for (let i = 0; i < size; i += 4) {
        seed = (seed * 1664525 + 1013904223) & 0x7fffffff;
        buffer.writeUInt32LE(seed, i);
    }

    // Encode to base64
    const base64Str = buffer.toString("base64");

    // Decode back to buffer
    const decoded = Buffer.from(base64Str, "base64");

    // Verify integrity
    const match = buffer.equals(decoded);
    console.log("Base64 transcode completed, bytes:", decoded.length, "match:", match);
}

run();
