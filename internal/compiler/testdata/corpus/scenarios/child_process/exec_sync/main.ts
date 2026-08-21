import { execSync } from "node:child_process";

const out1 = execSync("echo Hello from execSync");
console.log(out1.trim());

const out2 = execSync("echo 12345");
console.log(out2.trim());
