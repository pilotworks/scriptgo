import { Buffer } from "node:buffer";
const b = Buffer.alloc(4);
b.writeFloatLE(2.5, 0);
console.log(b.readFloatLE(0));
