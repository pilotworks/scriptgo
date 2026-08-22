// ScriptGo Corpus: Child_process Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as cp from "node:child_process";

// @api: child_process.execSync
// @expect: hello child
const out_child_process_execSync_0: string = cp.execSync("echo hello child");
console.log(out_child_process_execSync_0.trim());

// @api: child_process.spawnSync
// @expect: spawned
// @expect: 0
const ret_child_process_spawnSync_1 = cp.spawnSync("echo", ["spawned"]);
console.log(ret_child_process_spawnSync_1.stdout.trim());
console.log(ret_child_process_spawnSync_1.status);
