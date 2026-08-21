import { Buffer } from "node:buffer";
const b = Buffer.alloc(8);
b.writeDoubleLE(2.71828, 0);
console.log(b.readDoubleLE(0));
