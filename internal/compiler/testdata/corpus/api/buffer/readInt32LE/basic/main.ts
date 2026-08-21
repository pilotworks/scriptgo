import { Buffer } from "node:buffer";
const b = Buffer.alloc(4);
b.writeInt32LE(-500, 0);
console.log(b.readInt32LE(0));
