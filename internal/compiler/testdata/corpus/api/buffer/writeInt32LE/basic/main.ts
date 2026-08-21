import { Buffer } from "node:buffer";
const b = Buffer.alloc(4);
b.writeInt32LE(-12345, 0);
console.log(b.readInt32LE(0));
