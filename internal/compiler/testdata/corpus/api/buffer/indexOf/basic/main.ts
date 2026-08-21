import { Buffer } from "node:buffer";
const b = Buffer.from("hello world");
console.log(b.indexOf("world"));
console.log(b.indexOf("foo"));
