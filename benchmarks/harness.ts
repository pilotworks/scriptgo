// ScriptGo Benchmark Harness (Written in TypeScript, Executed natively by ScriptGo)

import { execSync } from "node:child_process";
import { performance } from "node:perf_hooks";
import { writeFileSync, mkdirSync } from "node:fs";
import { dirname } from "node:path";

interface BenchmarkCase {
    name: string;
    file: string;
    iterations: number;
}

interface MetricStats {
    medianMs: number;
    minMs: number;
    maxMs: number;
}

interface BenchmarkResult {
    name: string;
    scriptgo: MetricStats;
    node: MetricStats;
    bun: MetricStats | null;
    speedupVsNode: number;
    speedupVsBun: number | null;
}

const CASES: BenchmarkCase[] = [
    { name: "Cold Start Latency", file: "benchmarks/suites/cold_start.ts", iterations: 15 },
    { name: "Quicksort 100k", file: "benchmarks/suites/quicksort.ts", iterations: 7 },
    { name: "Matrix Mult 256x256", file: "benchmarks/suites/matrix_mult.ts", iterations: 5 },
    { name: "Buffer Ops (10MB)", file: "benchmarks/suites/buffer_ops.ts", iterations: 7 },
    { name: "Object Churn & GC", file: "benchmarks/suites/object_churn.ts", iterations: 5 },
];

function runCommandAndMeasure(cmd: string, runs: number): MetricStats {
    // Warmup round
    try {
        execSync(cmd);
    } catch {
        // ignore warmup failure
    }

    const times: number[] = [];
    for (let i = 0; i < runs; i++) {
        const start = performance.now();
        execSync(cmd);
        const duration = performance.now() - start;
        times.push(duration);
    }

    times.sort((a, b) => a - b);
    const medianMs = times[Math.floor(times.length / 2)];
    return {
        medianMs: Math.round(medianMs * 100) / 100,
        minMs: Math.round(times[0] * 100) / 100,
        maxMs: Math.round(times[times.length - 1] * 100) / 100,
    };
}

function padEnd(str: string, width: number): string {
    while (str.length < width) {
        str = str + " ";
    }
    return str;
}

function padStart(str: string, width: number): string {
    while (str.length < width) {
        str = " " + str;
    }
    return str;
}

function main(): void {
    console.log("====================================================================================================");
    console.log("               SCRIPTGO BENCHMARK HARNESS (AOT Native vs Node.js JIT vs Bun JIT)");
    console.log("====================================================================================================");

    // 1. Detect Node & Bun versions
    let nodeVer = "Node.js";
    try {
        nodeVer = "Node " + execSync("node -v").trim();
    } catch {}

    let hasBun = false;
    let bunVer = "Bun";
    try {
        const out = execSync("bun -v").trim();
        if (out.length > 0) {
            hasBun = true;
            bunVer = "Bun v" + out;
        }
    } catch {}

    // 2. Pre-compile ScriptGo CLI helper
    const scriptgoCli = "/tmp/scriptgo_bench_runner";
    console.log("==> Building native scriptgo compiler binary...");
    execSync("go build -o " + scriptgoCli + " ./cmd/scriptgo");
    console.log("==> Compiler ready at " + scriptgoCli + "\n");

    const header = padEnd("Benchmark", 22) +
        padStart("ScriptGo (AOT)", 16) +
        padStart(nodeVer, 16) +
        padStart(hasBun ? bunVer : "Bun", 16) +
        padStart("vs Node", 18) +
        padStart("vs Bun", 18);

    console.log(header);
    console.log("-".repeat(106));

    const results: BenchmarkResult[] = [];

    for (const testCase of CASES) {
        // Compile test case to optimized native executable
        const binTarget = "/tmp/sg_bench_" + testCase.name.replace(/[^a-zA-Z0-9]/g, "_");
        execSync(scriptgoCli + " build -O 3 --release " + testCase.file + " -o " + binTarget);

        // Measure ScriptGo native
        const sgStats = runCommandAndMeasure(binTarget, testCase.iterations);

        // Measure Node.js
        const nodeStats = runCommandAndMeasure("node " + testCase.file, testCase.iterations);

        // Measure Bun (if available)
        let bunStats: MetricStats | null = null;
        if (hasBun) {
            try {
                bunStats = runCommandAndMeasure("bun " + testCase.file, testCase.iterations);
            } catch {}
        }

        const speedupNode = Math.round((nodeStats.medianMs / sgStats.medianMs) * 100) / 100;
        let speedupBun: number | null = null;
        if (bunStats !== null && bunStats.medianMs > 0) {
            speedupBun = Math.round((bunStats.medianMs / sgStats.medianMs) * 100) / 100;
        }

        const sgStr = sgStats.medianMs.toFixed(1) + " ms";
        const nodeStr = nodeStats.medianMs.toFixed(1) + " ms";
        const bunStr = bunStats !== null ? bunStats.medianMs.toFixed(1) + " ms" : "N/A";
        const speedupNodeStr = (speedupNode >= 1.0 ? speedupNode.toFixed(2) + "x faster" : (1 / speedupNode).toFixed(2) + "x slower");
        let speedupBunStr = "N/A";
        if (speedupBun !== null) {
            speedupBunStr = (speedupBun >= 1.0 ? speedupBun.toFixed(2) + "x faster" : (1 / speedupBun).toFixed(2) + "x slower");
        }

        console.log(
            padEnd(testCase.name, 22) +
            padStart(sgStr, 16) +
            padStart(nodeStr, 16) +
            padStart(bunStr, 16) +
            padStart(speedupNodeStr, 18) +
            padStart(speedupBunStr, 18)
        );

        results.push({
            name: testCase.name,
            scriptgo: sgStats,
            node: nodeStats,
            bun: bunStats,
            speedupVsNode: speedupNode,
            speedupVsBun: speedupBun,
        });
    }

    console.log("=".repeat(106) + "\n");

    // Export results to JSON
    const reportPath = "web/src/data/benchmark-results.json";
    try {
        mkdirSync(dirname(reportPath), { recursive: true });
    } catch {}
    writeFileSync(reportPath, JSON.stringify(results, null, 2));
    console.log("✔ Benchmark report exported to " + reportPath);
}

main();
