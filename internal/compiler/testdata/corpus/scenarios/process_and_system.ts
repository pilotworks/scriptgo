// ScriptGo Corpus: Scenario: Process & System Execution
// Consolidated test suite with inline assertions.

import { execSync } from "node:child_process";
import { spawnSync } from "node:child_process";
import * as path from "node:path";
import * as os from "node:os";

// --- Context Case: scenarios_child_process_exec_sync ---
// @expect: Hello from execSync
// @expect: 12345
const out1_child_process_exec_sync_0 = execSync("echo Hello from execSync");
console.log(out1_child_process_exec_sync_0.trim());

const out2_child_process_exec_sync_0 = execSync("echo 12345");
console.log(out2_child_process_exec_sync_0.trim());

// --- Context Case: scenarios_child_process_spawn_sync ---
// @expect: Hello from spawnSync
// @expect: 0
const args_child_process_spawn_sync_1: string[] = ["Hello", "from", "spawnSync"];
const res_child_process_spawn_sync_1 = spawnSync("echo", args_child_process_spawn_sync_1);
console.log(res_child_process_spawn_sync_1.stdout.trim());
console.log(res_child_process_spawn_sync_1.status);

// --- Context Case: scenarios_modules_node_prefix ---
// @expect: foo/bar/baz.txt
// @expect: /usr/local
// @expect: bin
// @expect: .pdf
console.log(path.join("foo", "bar", "baz.txt"));
console.log(path.dirname("/usr/local/bin"));
console.log(path.basename("/usr/local/bin"));
console.log(path.extname("report.pdf"));

// --- Context Case: scenarios_os_node_prefix ---
// @expect: true
// @expect: true
// @expect: true
console.log(os.platform().length > 0);
console.log(os.arch().length > 0);
console.log(os.homedir().length > 0);

// --- Context Case: scenarios_process_global ---
// @expect: true
// @expect: true
const args_process_global_4: string[] = process.argv;
console.log(args_process_global_4.length > 0);
console.log(process.cwd().length > 0);
