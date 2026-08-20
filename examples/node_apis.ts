// Node.js Standard Library APIs Demo: node:path, node:crypto

import * as path from "node:path";
import * as crypto from "node:crypto";

console.log("=== Node.js Standard Library Demo ===");

// 1. Path operations
const fullPath: string = path.join("/usr", "local", "bin", "scriptgo");
console.log(`Path join: ${fullPath}`);
console.log(`Path dirname: ${path.dirname(fullPath)}`);
console.log(`Path basename: ${path.basename(fullPath)}`);
console.log(`Path extname: ${path.extname(fullPath)}`);

// 2. Crypto operations
const uuid: string = crypto.randomUUID();
console.log(`Random UUID: ${uuid}`);
console.log(`Random UUID length: ${uuid.length}`);
console.log(`UUID has hyphen: ${uuid.includes("-")}`);
