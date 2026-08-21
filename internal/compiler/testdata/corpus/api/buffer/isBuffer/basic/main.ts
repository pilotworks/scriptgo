import { Buffer } from "node:buffer"; const b = Buffer.alloc(2); console.log(Buffer.isBuffer(b)); console.log(Buffer.isBuffer("str"));
