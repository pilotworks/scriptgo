import { Buffer } from "node:buffer";

const a: Buffer = Buffer.from("hello ");
const b: Buffer = Buffer.from("world");
const c: Buffer = Buffer.concat([a, b]);
console.log(c.toString());

console.log(Buffer.isBuffer(c));
console.log(Buffer.isBuffer("not a buffer"));

const d: Buffer = Buffer.alloc(11);
c.copy(d, 0, 0, 11);
console.log(d.toString());

console.log(c.equals(d));
console.log(c.compare(d));

const idx: number = c.indexOf("world");
console.log(idx);

const sub: Buffer = c.subarray(6, 11);
console.log(sub.toString());
