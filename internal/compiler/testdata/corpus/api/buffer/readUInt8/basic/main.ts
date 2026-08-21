import { Buffer } from "node:buffer";
const b = Buffer.from("A");
console.log(b.readUInt8(0));
