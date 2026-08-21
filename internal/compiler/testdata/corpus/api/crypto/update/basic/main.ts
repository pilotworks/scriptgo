import { createHash } from "node:crypto";
const h = createHash("sha256");
h.update("hello");
console.log(h.digest("hex").length > 0);
