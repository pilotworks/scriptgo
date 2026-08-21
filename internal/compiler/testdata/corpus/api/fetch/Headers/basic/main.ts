import { Headers } from "node:http";
const h = new Headers();
h.set("Content-Type", "application/json");
console.log(h.get("Content-Type"));
console.log(h.has("Content-Type"));
