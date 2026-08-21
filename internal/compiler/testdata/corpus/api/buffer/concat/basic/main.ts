import { Buffer } from "node:buffer";
const b1 = Buffer.from("hello ");
const b2 = Buffer.from("world");
const b3 = Buffer.concat([b1, b2]);
console.log(b3.toString("utf8"));
