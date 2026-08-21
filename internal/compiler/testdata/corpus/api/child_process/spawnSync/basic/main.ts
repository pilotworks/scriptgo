import * as cp from "node:child_process";
const ret = cp.spawnSync("echo", ["spawned"]);
console.log(ret.stdout.trim());
console.log(ret.status);
