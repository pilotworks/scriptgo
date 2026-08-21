import { Buffer } from "node:buffer";
const b = Buffer.alloc(1);
b.writeUInt8(66, 0);
console.log(b.toString("utf8"));
