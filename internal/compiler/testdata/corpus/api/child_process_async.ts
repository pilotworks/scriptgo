// @expect: === Testing spawn async ===
// @expect: child spawned, pid > 0: true
// @expect: child exited with code: 0
// @expect: child closed with output: async_spawn_success
// @expect: === Testing exec ===
// @expect: exec err is null: true
// @expect: exec stdout: exec_success
// @expect: === Testing execFile ===
// @expect: execFile err is null: true
// @expect: execFile stdout: exec_file_success
// @expect: ALL CHILD PROCESS ASYNC TESTS COMPLETED

import { spawn, exec, execFile } from "node:child_process";

console.log("=== Testing spawn async ===");

const child = spawn("echo", ["async_spawn_success"]);
let childOutput = "";

if (child.stdout !== null) {
    child.stdout.on("data", (chunk: string) => {
        childOutput += String(chunk);
    });
}

child.on("spawn", () => {
    console.log("child spawned, pid > 0:", child.pid > 0);
});

child.on("exit", (code: number) => {
    console.log("child exited with code:", code);
});

child.on("close", (code: number) => {
    console.log("child closed with output:", childOutput.trim());

    console.log("=== Testing exec ===");
    exec("echo exec_success", (err: Error | null, stdout: string, stderr: string) => {
        console.log("exec err is null:", err === null);
        console.log("exec stdout:", stdout.trim());

        console.log("=== Testing execFile ===");
        execFile("echo", ["exec_file_success"], (err2: Error | null, stdout2: string, stderr2: string) => {
            console.log("execFile err is null:", err2 === null);
            console.log("execFile stdout:", stdout2.trim());
            console.log("ALL CHILD PROCESS ASYNC TESTS COMPLETED");
        });
    });
});
