import { Buffer } from "node:buffer";
const b = Buffer.alloc(8);
b.writeDoubleLE(3.14159, 0);
console.log(b.readDoubleLE(0));
