import * as crypto from "node:crypto";
const hash = crypto.createHash("sha256").update("hello").digest("hex");
console.log(hash.length === 64);
