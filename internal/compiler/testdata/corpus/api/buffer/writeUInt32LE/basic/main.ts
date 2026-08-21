import { Buffer } from "node:buffer";
const b = Buffer.alloc(4);
b.writeUInt32LE(999999, 0);
console.log(b.readUInt32LE(0));
