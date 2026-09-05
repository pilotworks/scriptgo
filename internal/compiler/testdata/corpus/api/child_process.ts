// ScriptGo Corpus: Child_process Standard Builtin APIs
// Tests real synchronous execution backed by C runtime POSIX fork/exec/popen.

import {
    spawnSync,
    execSync,
    execFileSync,
} from "node:child_process";

// @api: child_process.execSync
// @expect: execSync_res: hello
const execSyncOut = execSync("echo hello");
console.log("execSync_res: " + execSyncOut.toString().trim());

// @api: child_process.spawnSync
// @expect: spawnSync_res: hello
// @expect: spawnSync_status: 0
const spawnSyncRet = spawnSync("echo", ["hello"]);
console.log("spawnSync_res: " + spawnSyncRet.stdout.toString().trim());
console.log("spawnSync_status: " + spawnSyncRet.status);

// @api: child_process.execFileSync
// @expect: execFileSync_res: hello
const execFileSyncOut = execFileSync("echo", ["hello"]);
console.log("execFileSync_res: " + execFileSyncOut.toString().trim());
