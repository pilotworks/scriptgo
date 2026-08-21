import * as crypto from "crypto";

const id1: string = crypto.randomUUID();
console.log(id1.length);
console.log(id1.indexOf("-"));
