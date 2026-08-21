import { spawnSync } from "node:child_process";

const args: string[] = ["Hello", "from", "spawnSync"];
const res = spawnSync("echo", args);
console.log(res.stdout.trim());
console.log(res.status);
