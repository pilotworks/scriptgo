import * as cp from "node:child_process";
const out: string = cp.execSync("echo hello child");
console.log(out.trim());
