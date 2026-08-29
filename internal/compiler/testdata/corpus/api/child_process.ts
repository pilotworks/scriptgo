// ScriptGo Corpus: Child_process Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions for all 26 official Node.js child_process APIs.

import {
    ChildProcess,
    spawn,
    exec,
    execFile,
    fork,
    spawnSync,
    execSync,
    execFileSync
} from "node:child_process";

// @api: new child_process.ChildProcess
// @expect: cp_inst: true
const proc = new ChildProcess("echo", ["hi"]);
console.log("cp_inst: " + (proc instanceof ChildProcess));

// @api: ChildProcess.pid
// @expect: cp_pid: true
console.log("cp_pid: " + (proc.pid >= 0));

// @api: ChildProcess.exitCode
// @expect: cp_exitCode: true
console.log("cp_exitCode: " + (proc.exitCode === 0));

// @api: ChildProcess.signalCode
// @expect: cp_signalCode: true
console.log("cp_signalCode: " + (proc.signalCode === null));

// @api: ChildProcess.spawnfile
// @expect: cp_spawnfile: echo
console.log("cp_spawnfile: " + proc.spawnfile);

// @api: ChildProcess.spawnargs
// @expect: cp_spawnargs: true
console.log("cp_spawnargs: " + (proc.spawnargs.length === 1));

// @api: ChildProcess.killed
// @expect: cp_killed: false
console.log("cp_killed: " + proc.killed);

// @api: ChildProcess.connected
// @expect: cp_connected: false
console.log("cp_connected: " + proc.connected);

// @api: ChildProcess.channel
// @expect: cp_channel: true
console.log("cp_channel: " + (proc.channel === null));

// @api: ChildProcess.stdin
// @expect: cp_stdin: true
console.log("cp_stdin: " + (proc.stdin === null));

// @api: ChildProcess.stdout
// @expect: cp_stdout: true
console.log("cp_stdout: " + (proc.stdout === null));

// @api: ChildProcess.stderr
// @expect: cp_stderr: true
console.log("cp_stderr: " + (proc.stderr === null));

// @api: ChildProcess.stdio
// @expect: cp_stdio: true
console.log("cp_stdio: " + Array.isArray(proc.stdio));

// @api: ChildProcess.ref
// @expect: cp_ref: true
console.log("cp_ref: " + (proc.ref() === proc));

// @api: ChildProcess.unref
// @expect: cp_unref: true
console.log("cp_unref: " + (proc.unref() === proc));

// @api: ChildProcess.send
// @expect: cp_send: true
console.log("cp_send: " + proc.send("msg"));

// @api: ChildProcess.disconnect
// @expect: cp_disconnect: true
proc.disconnect();
console.log("cp_disconnect: true");

// @api: ChildProcess.kill
// @expect: cp_kill: true
console.log("cp_kill: " + proc.kill());

// @api: ChildProcess.[Symbol.dispose]
// @expect: cp_dispose: true
proc[Symbol.dispose]();
console.log("cp_dispose: true");

// @api: child_process.spawn
// @expect: spawn_res: true
const spawned = spawn("echo", ["hello"]);
console.log("spawn_res: " + (spawned instanceof ChildProcess));

// @api: child_process.exec
// @expect: exec_res: true
const executed = exec("echo hello");
console.log("exec_res: " + (executed instanceof ChildProcess));

// @api: child_process.execFile
// @expect: execFile_res: true
const execFiled = execFile("echo", ["hello"]);
console.log("execFile_res: " + (execFiled instanceof ChildProcess));

// @api: child_process.fork
// @expect: fork_res: true
const forked = fork("module.js");
console.log("fork_res: " + (forked instanceof ChildProcess));

// @api: child_process.execSync
// @expect: execSync_res: hello
const execSyncOut = execSync("echo hello");
console.log("execSync_res: " + execSyncOut.trim());

// @api: child_process.spawnSync
// @expect: spawnSync_res: hello
// @expect: spawnSync_status: 0
const spawnSyncRet = spawnSync("echo", ["hello"]);
console.log("spawnSync_res: " + spawnSyncRet.stdout.trim());
console.log("spawnSync_status: " + spawnSyncRet.status);

// @api: child_process.execFileSync
// @expect: execFileSync_res: hello
const execFileSyncOut = execFileSync("echo hello");
console.log("execFileSync_res: " + execFileSyncOut.trim());

