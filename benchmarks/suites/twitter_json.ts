// Twitter JSON Benchmark: Real-world complex hierarchical JSON dataset (617KB)
// Tests deep object hierarchies, UTF-8 strings, and dynamic property serialization.
import { readFileSync, existsSync } from "node:fs";

function resolveDataset(): string {
    const candidates = [
        "benchmarks/data/twitter.json",
        "data/twitter.json",
        "../benchmarks/data/twitter.json",
    ];
    for (let i = 0; i < candidates.length; i++) {
        if (existsSync(candidates[i])) {
            return candidates[i];
        }
    }
    return "benchmarks/data/twitter.json";
}

function run(): void {
    const filePath = resolveDataset();
    const raw = readFileSync(filePath, "utf-8");
    const iterations = 5;
    let checksum = 0;

    for (let i = 0; i < iterations; i++) {
        const parsed: unknown = JSON.parse(raw);
        const serialized = JSON.stringify(parsed);
        checksum += serialized.length;
    }

    console.log("Twitter benchmark completed, bytes:", raw.length, "iterations:", iterations, "checksum:", checksum);
}

run();
