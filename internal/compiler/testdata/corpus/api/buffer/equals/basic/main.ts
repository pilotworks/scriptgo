import { Buffer } from "node:buffer";
const b1 = Buffer.from("abc");
const b2 = Buffer.from("abc");
const b3 = Buffer.from("def");
console.log(b1.equals(b2));
console.log(b1.equals(b3));
