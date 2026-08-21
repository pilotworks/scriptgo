import { Buffer } from "node:buffer";
const src = Buffer.from("hello");
const dst = Buffer.alloc(5);
src.copy(dst);
console.log(dst.toString("utf8"));
