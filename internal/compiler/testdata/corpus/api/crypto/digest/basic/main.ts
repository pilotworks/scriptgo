import { createHash } from "node:crypto";
const h = createHash("sha256");
h.update("test");
const d: string = h.digest("hex");
console.log(d.length > 0);
