import * as crypto from "node:crypto";
const uuid: string = crypto.randomUUID();
console.log(uuid.length > 0);
