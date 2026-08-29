package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmTargetBuildAndRun(t *testing.T) {
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig is not installed (required for wasm32-wasi cross-compilation)")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed (required for executing wasm32-wasi binaries)")
	}

	dir := t.TempDir()
	entry := filepath.Join(dir, "main.ts")
	outputWasm := filepath.Join(dir, "main.wasm")

	tsCode := `
console.log("Hello WebAssembly from ScriptGo!");

function fib(n: number): number {
    if (n <= 1) return n;
    return fib(n - 1) + fib(n - 2);
}

console.log("fib(10) = " + fib(10));

const arr: number[] = [10, 20, 30];
let sum = 0;
for (let i = 0; i < arr.length; i++) {
    sum += arr[i];
}
console.log("sum = " + sum);
`
	if err := os.WriteFile(entry, []byte(tsCode), 0o644); err != nil {
		t.Fatal(err)
	}

	options := BuildOptions{
		Target: "wasm32-wasi",
	}

	if err := BuildWithOptions(entry, outputWasm, options); err != nil {
		t.Fatalf("BuildWithOptions(wasm32-wasi) failed: %v", err)
	}

	info, err := os.Stat(outputWasm)
	if err != nil {
		t.Fatalf("failed to stat generated wasm file: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("generated wasm file is empty")
	}

	// Execute generated WASM binary using Node.js WASI API
	nodeRunner := `
const { WASI } = require("wasi");
const fs = require("fs");
const wasi = new WASI({
    version: "preview1",
    args: [process.argv[2]],
    env: process.env,
    returnOnExit: true
});
const wasmBytes = fs.readFileSync(process.argv[2]);
(async () => {
    try {
        const mod = await WebAssembly.compile(wasmBytes);
        const inst = await WebAssembly.instantiate(mod, wasi.getImportObject());
        const exitCode = wasi.start(inst);
        process.exit(exitCode || 0);
    } catch (e) {
        console.error("WASI runtime error:", e);
        process.exit(1);
    }
})();
`
	runnerPath := filepath.Join(dir, "runner.js")
	if err := os.WriteFile(runnerPath, []byte(nodeRunner), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", runnerPath, outputWasm)
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	if err != nil {
		t.Fatalf("node WASI execution failed: %v\nOutput:\n%s", err, outStr)
	}

	if !strings.Contains(outStr, "Hello WebAssembly from ScriptGo!") {
		t.Errorf("missing expected greeting in wasm output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "fib(10) = 55") {
		t.Errorf("missing expected fib result in wasm output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "sum = 60") {
		t.Errorf("missing expected sum in wasm output:\n%s", outStr)
	}
}

func TestWasmComplexFeatures(t *testing.T) {
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig is not installed (required for wasm32-wasi cross-compilation)")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed (required for executing wasm32-wasi binaries)")
	}

	dir := t.TempDir()
	entry := filepath.Join(dir, "complex.ts")
	outputWasm := filepath.Join(dir, "complex.wasm")

	tsCode := `
class Greeter {
    greeting: string;
    constructor(message: string) {
        this.greeting = message;
    }
    greet(): string {
        return "Greeting: " + this.greeting;
    }
}

const g = new Greeter("ScriptGo on WASI");
console.log(g.greet());

function makeCounter(init: number): () => number {
    let count = init;
    return () => {
        count += 1;
        return count;
    };
}

const c = makeCounter(100);
console.log("c1 = " + c());
console.log("c2 = " + c());

const words: string[] = ["web", "assembly", "is", "fast"];
const joined = words.join("-");
console.log("joined = " + joined);
`
	if err := os.WriteFile(entry, []byte(tsCode), 0o644); err != nil {
		t.Fatal(err)
	}

	options := BuildOptions{
		Target: "wasm32-wasi",
	}

	if err := BuildWithOptions(entry, outputWasm, options); err != nil {
		t.Fatalf("BuildWithOptions(wasm32-wasi) failed: %v", err)
	}

	nodeRunner := `
const { WASI } = require("wasi");
const fs = require("fs");
const wasi = new WASI({
    version: "preview1",
    args: [process.argv[2]],
    env: process.env,
    returnOnExit: true
});
const wasmBytes = fs.readFileSync(process.argv[2]);
(async () => {
    try {
        const mod = await WebAssembly.compile(wasmBytes);
        const inst = await WebAssembly.instantiate(mod, wasi.getImportObject());
        const exitCode = wasi.start(inst);
        process.exit(exitCode || 0);
    } catch (e) {
        console.error("WASI runtime error:", e);
        process.exit(1);
    }
})();
`
	runnerPath := filepath.Join(dir, "runner.js")
	if err := os.WriteFile(runnerPath, []byte(nodeRunner), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", runnerPath, outputWasm)
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	if err != nil {
		t.Fatalf("node WASI execution failed: %v\nOutput:\n%s", err, outStr)
	}

	if !strings.Contains(outStr, "Greeting: ScriptGo on WASI") {
		t.Errorf("missing expected class greeting in wasm output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "c1 = 101") || !strings.Contains(outStr, "c2 = 102") {
		t.Errorf("missing expected closure counter in wasm output:\n%s", outStr)
	}
	if !strings.Contains(outStr, "joined = web-assembly-is-fast") {
		t.Errorf("missing expected joined string in wasm output:\n%s", outStr)
	}
}

func TestWasmStdlibAndJson(t *testing.T) {
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig is not installed (required for wasm32-wasi cross-compilation)")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed (required for executing wasm32-wasi binaries)")
	}

	dir := t.TempDir()
	entry := filepath.Join(dir, "stdlib.ts")
	outputWasm := filepath.Join(dir, "stdlib.wasm")

	tsCode := `
const pi = Math.PI;
const rounded = Math.round(pi * 100) / 100;
console.log("PI approx = " + rounded);

const str = "ScriptGo WebAssembly Target";
console.log("upper = " + str.toUpperCase());
console.log("includes WASM = " + (str.includes("WebAssembly") ? "true" : "false"));

const data = { name: "ScriptGo", version: 1, targets: ["native", "wasm32-wasi"] };
const jsonStr = JSON.stringify(data);
console.log("JSON = " + jsonStr);
`
	if err := os.WriteFile(entry, []byte(tsCode), 0o644); err != nil {
		t.Fatal(err)
	}

	options := BuildOptions{
		Target: "wasm32-wasi",
	}

	if err := BuildWithOptions(entry, outputWasm, options); err != nil {
		t.Fatalf("BuildWithOptions(wasm32-wasi) failed: %v", err)
	}

	nodeRunner := `
const { WASI } = require("wasi");
const fs = require("fs");
const wasi = new WASI({
    version: "preview1",
    args: [process.argv[2]],
    env: process.env,
    returnOnExit: true
});
const wasmBytes = fs.readFileSync(process.argv[2]);
(async () => {
    try {
        const mod = await WebAssembly.compile(wasmBytes);
        const inst = await WebAssembly.instantiate(mod, wasi.getImportObject());
        const exitCode = wasi.start(inst);
        process.exit(exitCode || 0);
    } catch (e) {
        console.error("WASI runtime error:", e);
        process.exit(1);
    }
})();
`
	runnerPath := filepath.Join(dir, "runner.js")
	if err := os.WriteFile(runnerPath, []byte(nodeRunner), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", runnerPath, outputWasm)
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	if err != nil {
		t.Fatalf("node WASI execution failed: %v\nOutput:\n%s", err, outStr)
	}

	if !strings.Contains(outStr, "PI approx = 3.14") {
		t.Errorf("missing expected Math result in wasm output:\n%s", outStr)
	}
	if !strings.Contains(outStr, `"name":"ScriptGo"`) {
		t.Errorf("missing expected JSON in wasm output:\n%s", outStr)
	}
}

func TestWasmCorpusFunctionsAndClosures(t *testing.T) {
	if _, err := exec.LookPath("zig"); err != nil {
		t.Skip("zig is not installed (required for wasm32-wasi cross-compilation)")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed (required for executing wasm32-wasi binaries)")
	}

	dir := t.TempDir()
	corpusEntry := filepath.Join("testdata", "corpus", "language", "functions_and_closures.ts")
	outputWasm := filepath.Join(dir, "app.wasm")

	options := BuildOptions{
		Target: "wasm32-wasi",
	}

	if err := BuildWithOptions(corpusEntry, outputWasm, options); err != nil {
		t.Fatalf("BuildWithOptions(wasm32-wasi) failed on corpus test: %v", err)
	}

	nodeRunner := `
const { WASI } = require("wasi");
const fs = require("fs");
const wasi = new WASI({
    version: "preview1",
    args: [process.argv[2]],
    env: process.env,
    returnOnExit: true
});
const wasmBytes = fs.readFileSync(process.argv[2]);
(async () => {
    try {
        const mod = await WebAssembly.compile(wasmBytes);
        const inst = await WebAssembly.instantiate(mod, wasi.getImportObject());
        const exitCode = wasi.start(inst);
        process.exit(exitCode || 0);
    } catch (e) {
        console.error("WASI runtime error:", e);
        process.exit(1);
    }
})();
`
	runnerPath := filepath.Join(dir, "runner.js")
	if err := os.WriteFile(runnerPath, []byte(nodeRunner), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("node", runnerPath, outputWasm)
	outBytes, err := cmd.CombinedOutput()
	outStr := string(outBytes)
	if err != nil {
		t.Fatalf("node WASI execution failed: %v\nOutput:\n%s", err, outStr)
	}

	expectedSnippets := []string{
		"2, 4, 6, 8, 10",
		"2, 4",
		"Hello, ScriptGo!",
		"USD $100 total",
		"USD $250.5 total",
		"EUR 90 net",
		"JPY 1500 YEN",
		"[WARN] Server starting on port 8080",
		"[ERROR] Database connection lost",
		"start:a:b:c",
	}
	for _, snippet := range expectedSnippets {
		if !strings.Contains(outStr, snippet) {
			t.Errorf("missing expected snippet %q in wasm output:\n%s", snippet, outStr)
		}
	}
}
