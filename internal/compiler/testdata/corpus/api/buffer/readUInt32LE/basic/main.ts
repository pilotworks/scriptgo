import { Buffer } from "node:buffer";
const b = Buffer.alloc(4);
b.writeUInt32LE(123456, 0);
console.log(b.readUInt32LE(0));
