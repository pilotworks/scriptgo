import { URLSearchParams } from "node:url";
const p = new URLSearchParams("a=1&b=2");
console.log(p.get("a"));
console.log(p.get("b"));
