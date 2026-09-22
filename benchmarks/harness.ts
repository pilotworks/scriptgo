// ScriptGo Benchmark Harness (Written in TypeScript, Executed natively by ScriptGo)

import { execSync } from "node:child_process";
import { performance } from "node:perf_hooks";
import { writeFileSync, mkdirSync, statSync, existsSync } from "node:fs";
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

interface BenchmarkTimeResult {
    scriptgo: MetricStats;
    node: MetricStats;
    bun: MetricStats | null;
    speedupVsNode: number;
    speedupVsBun: number | null;
}

interface BenchmarkMemoryResult {
    scriptgoRssMb: number;
    nodeRssMb: number;
    bunRssMb: number | null;
    ramSavingsVsNode: number;
    ramSavingsVsBun: number | null;
}

interface BenchmarkBuildResult {
    binarySizeBytes: number;
    binarySizeFormatted: string;
    compileTimeMs: number;
}

interface BenchmarkResult {
    name: string;
    time: BenchmarkTimeResult;
    memory: BenchmarkMemoryResult;
    build: BenchmarkBuildResult;
}

const CASES: BenchmarkCase[] = [
    { name: "Cold Start Latency", file: "benchmarks/suites/cold_start.ts", iterations: 15 },
    { name: "Quicksort 100k", file: "benchmarks/suites/quicksort.ts", iterations: 7 },
    { name: "Matrix Mult 256x256", file: "benchmarks/suites/matrix_mult.ts", iterations: 5 },
    { name: "Buffer Ops (10MB)", file: "benchmarks/suites/buffer_ops.ts", iterations: 7 },
    { name: "Object Churn & GC", file: "benchmarks/suites/object_churn.ts", iterations: 5 },
    { name: "Mandelbrot 500x500", file: "benchmarks/suites/mandelbrot.ts", iterations: 5 },
    { name: "Binary Trees D14", file: "benchmarks/suites/binary_trees.ts", iterations: 5 },
    { name: "ES2024 Set Ops", file: "benchmarks/suites/es2024_set.ts", iterations: 5 },
    { name: "Base64 Transcode", file: "benchmarks/suites/base64_transcode.ts", iterations: 5 },
    { name: "JSON Ops (50 Records)", file: "benchmarks/suites/json_ops.ts", iterations: 5 },
    { name: "Twitter JSON (617KB)", file: "benchmarks/suites/twitter_json.ts", iterations: 5 },
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

function measurePeakRss(cmd: string): number {
    try {
        let timeCmd = "/usr/bin/time -l " + cmd + " 2>&1";
        if (process.platform === "linux") {
            timeCmd = "/usr/bin/time -v " + cmd + " 2>&1";
        }
        const out = String(execSync(timeCmd, { encoding: "utf-8" }));
        const lines = out.split("\n");
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            // macOS: "6291456  maximum resident set size"
            if (line.indexOf("maximum resident set size") !== -1) {
                const parts = line.split(" ");
                const bytes = parseFloat(parts[0]);
                if (!isNaN(bytes) && bytes > 0) {
                    return Math.round((bytes / (1024 * 1024)) * 10) / 10;
                }
            }
            // Linux: "Maximum resident set size (kbytes): 6144"
            if (line.indexOf("Maximum resident set size (kbytes):") !== -1) {
                const parts = line.split(":");
                if (parts.length > 1) {
                    const kbytes = parseFloat(parts[1].trim());
                    if (!isNaN(kbytes) && kbytes > 0) {
                        return Math.round((kbytes / 1024) * 10) / 10;
                    }
                }
            }
        }
    } catch {}
    return 0;
}

function formatBytes(bytes: number): string {
    if (bytes >= 1024 * 1024) {
        return (bytes / (1024 * 1024)).toFixed(2) + " MB";
    }
    return Math.round(bytes / 1024) + " KB";
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
    console.log("          SCRIPTGO MULTI-DIMENSIONAL BENCHMARK HARNESS (Speed, Memory & Binary Footprint)");
    console.log("====================================================================================================");

    // 1. Detect Node & Bun versions
    let nodeVer = "Node.js";
    try {
        nodeVer = "Node " + String(execSync("node -v", { encoding: "utf-8" })).trim();
    } catch {}

    let hasBun = false;
    let bunVer = "Bun";
    try {
        const out = String(execSync("bun -v", { encoding: "utf-8" })).trim();
        if (out.length > 0) {
            hasBun = true;
            bunVer = "Bun v" + out;
        }
    } catch {}

    // 2. Pre-compile ScriptGo CLI helper
    const repoRoot = existsSync("cmd/scriptgo") ? "." : (existsSync("../cmd/scriptgo") ? ".." : ".");
    const scriptgoCli = "/tmp/scriptgo_bench_runner";
    console.log("==> Building native scriptgo compiler binary...");
    execSync("go build -o " + scriptgoCli + " " + repoRoot + "/cmd/scriptgo");
    console.log("==> Compiler ready at " + scriptgoCli + "\n");

    const results: BenchmarkResult[] = [];

    console.log("==> Measuring " + CASES.length + " suites across ScriptGo (AOT Native), " + nodeVer + ", and " + (hasBun ? bunVer : "Bun") + "...\n");

    for (const testCase of CASES) {
        const suiteFile = existsSync(testCase.file) ? testCase.file : (existsSync(repoRoot + "/" + testCase.file) ? (repoRoot + "/" + testCase.file) : testCase.file.replace(/^benchmarks\//, ""));
        // Compile test case to optimized native executable
        const binTarget = "/tmp/sg_bench_" + testCase.name.replace(/[^a-zA-Z0-9]/g, "_");
        const compileStart = performance.now();
        execSync(scriptgoCli + " build -O 3 --release " + suiteFile + " -o " + binTarget);
        const compileTimeMs = Math.round(performance.now() - compileStart);
        let binarySizeBytes = 0;
        try {
            binarySizeBytes = statSync(binTarget).size;
        } catch {}

        // Measure execution times
        const sgStats = runCommandAndMeasure(binTarget, testCase.iterations);
        const nodeStats = runCommandAndMeasure("node " + suiteFile, testCase.iterations);
        let bunStats: MetricStats | null = null;
        if (hasBun) {
            try {
                bunStats = runCommandAndMeasure("bun " + suiteFile, testCase.iterations);
            } catch {}
        }

        const speedupNode = Math.round((nodeStats.medianMs / sgStats.medianMs) * 100) / 100;
        let speedupBun: number | null = null;
        if (bunStats !== null && bunStats.medianMs > 0) {
            speedupBun = Math.round((bunStats.medianMs / sgStats.medianMs) * 100) / 100;
        }

        // Measure Peak RSS
        const sgRssMb = measurePeakRss(binTarget);
        const nodeRssMb = measurePeakRss("node " + testCase.file);
        let bunRssMb: number | null = null;
        if (hasBun) {
            try {
                bunRssMb = measurePeakRss("bun " + testCase.file);
            } catch {}
        }

        let ramSavingsVsNode = 1.0;
        if (sgRssMb > 0 && nodeRssMb > 0) {
            ramSavingsVsNode = Math.round((nodeRssMb / sgRssMb) * 100) / 100;
        }

        let ramSavingsVsBun: number | null = null;
        if (hasBun && bunRssMb !== null && bunRssMb > 0 && sgRssMb > 0) {
            ramSavingsVsBun = Math.round((bunRssMb / sgRssMb) * 100) / 100;
        }

        results.push({
            name: testCase.name,
            time: {
                scriptgo: sgStats,
                node: nodeStats,
                bun: bunStats,
                speedupVsNode: speedupNode,
                speedupVsBun: speedupBun,
            },
            memory: {
                scriptgoRssMb: sgRssMb,
                nodeRssMb: nodeRssMb,
                bunRssMb: bunRssMb,
                ramSavingsVsNode: ramSavingsVsNode,
                ramSavingsVsBun: ramSavingsVsBun,
            },
            build: {
                binarySizeBytes: binarySizeBytes,
                binarySizeFormatted: formatBytes(binarySizeBytes),
                compileTimeMs: compileTimeMs,
            },
        });
    }

    // TABLE 1: EXECUTION SPEED & LATENCY
    console.log("┌" + "─".repeat(104) + "┐");
    console.log("│ " + padEnd("DIMENSION 1: EXECUTION LATENCY & SPEEDUP (Wall-Clock, lower is better)", 103) + "│");
    console.log("├" + "─".repeat(104) + "┤");
    const header1 = "│ " + padEnd("Benchmark", 22) +
        padStart("ScriptGo (AOT)", 16) +
        padStart(nodeVer, 15) +
        padStart(hasBun ? bunVer : "Bun", 15) +
        padStart("vs Node", 18) +
        padStart("vs Bun", 16) + " │";
    console.log(header1);
    console.log("├" + "─".repeat(104) + "┤");

    for (const r of results) {
        const sgStr = r.time.scriptgo.medianMs.toFixed(1) + " ms";
        const nodeStr = r.time.node.medianMs.toFixed(1) + " ms";
        let bunStr = "N/A";
        const bunTime = r.time.bun;
        if (bunTime !== null) {
            bunStr = bunTime.medianMs.toFixed(1) + " ms";
        }
        const speedupNodeStr = (r.time.speedupVsNode >= 1.0 ? r.time.speedupVsNode.toFixed(2) + "x faster" : (1 / r.time.speedupVsNode).toFixed(2) + "x slower");
        let speedupBunStr = "N/A";
        const speedupBunVal = r.time.speedupVsBun;
        if (speedupBunVal !== null) {
            speedupBunStr = (speedupBunVal >= 1.0 ? speedupBunVal.toFixed(2) + "x faster" : (1 / speedupBunVal).toFixed(2) + "x slower");
        }
        console.log(
            "│ " + padEnd(r.name, 22) +
            padStart(sgStr, 16) +
            padStart(nodeStr, 15) +
            padStart(bunStr, 15) +
            padStart(speedupNodeStr, 18) +
            padStart(speedupBunStr, 16) + " │"
        );
    }
    console.log("└" + "─".repeat(104) + "┘\n");

    // TABLE 2: MEMORY FOOTPRINT (PEAK RSS)
    console.log("┌" + "─".repeat(104) + "┐");
    console.log("│ " + padEnd("DIMENSION 2: MEMORY FOOTPRINT (Peak Resident Set Size, lower is better)", 103) + "│");
    console.log("├" + "─".repeat(104) + "┤");
    const header2 = "│ " + padEnd("Benchmark", 22) +
        padStart("ScriptGo RSS", 16) +
        padStart("Node RSS", 15) +
        padStart("Bun RSS", 15) +
        padStart("RAM vs Node", 18) +
        padStart("RAM vs Bun", 16) + " │";
    console.log(header2);
    console.log("├" + "─".repeat(104) + "┤");

    for (const r of results) {
        const sgRss = r.memory.scriptgoRssMb.toFixed(1) + " MB";
        const nodeRss = r.memory.nodeRssMb.toFixed(1) + " MB";
        let bunRss = "N/A";
        const bunRssVal = r.memory.bunRssMb;
        if (bunRssVal !== null) {
            bunRss = bunRssVal.toFixed(1) + " MB";
        }
        const ramNodeStr = r.memory.ramSavingsVsNode.toFixed(1) + "x less RAM";
        let ramBunStr = "N/A";
        const ramBunVal = r.memory.ramSavingsVsBun;
        if (ramBunVal !== null) {
            ramBunStr = ramBunVal.toFixed(1) + "x less RAM";
        }
        console.log(
            "│ " + padEnd(r.name, 22) +
            padStart(sgRss, 16) +
            padStart(nodeRss, 15) +
            padStart(bunRss, 15) +
            padStart(ramNodeStr, 18) +
            padStart(ramBunStr, 16) + " │"
        );
    }
    console.log("└" + "─".repeat(104) + "┘\n");

    // TABLE 3: ARTIFACT SIZE & COMPILE TIME
    console.log("┌" + "─".repeat(104) + "┐");
    console.log("│ " + padEnd("DIMENSION 3: AOT COMPILATION & STANDALONE BINARY FOOTPRINT", 103) + "│");
    console.log("├" + "─".repeat(104) + "┤");
    const header3 = "│ " + padEnd("Benchmark Suite", 30) +
        padStart("Standalone Binary Size", 26) +
        padStart("AOT Compile Duration", 24) +
        padStart("Target Optimization", 22) + " │";
    console.log(header3);
    console.log("├" + "─".repeat(104) + "┤");

    for (const r of results) {
        console.log(
            "│ " + padEnd(r.name, 30) +
            padStart(r.build.binarySizeFormatted, 26) +
            padStart(r.build.compileTimeMs + " ms", 24) +
            padStart("-O3 --release --lto", 22) + " │"
        );
    }
    console.log("└" + "─".repeat(104) + "┘\n");

    // Export results to JSON
    const reportPath = repoRoot + "/web/src/data/benchmark-results.json";
    try {
        mkdirSync(dirname(reportPath), { recursive: true });
    } catch {}
    writeFileSync(reportPath, JSON.stringify(results, null, 2));
    console.log("✔ Multi-dimensional benchmark report exported to " + reportPath);
}

main();
