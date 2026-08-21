import { Buffer } from "node:buffer";
const b = Buffer.alloc(2);
b.writeUInt16LE(0x1234, 0);
console.log(b.readUInt16LE(0));
