import { Buffer } from "node:buffer";
const b1 = Buffer.from("abc");
const b2 = Buffer.from("abc");
console.log(b1.compare(b2));
