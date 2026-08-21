import { Headers } from "node:http";

const h = new Headers();
h.set("Content-Type", "application/json");
h.append("Accept", "text/plain");
h.append("accept", "text/html");

console.log(h.get("content-type"));
console.log(h.get("Accept"));
console.log(h.has("CONTENT-TYPE"));
console.log(h.has("Authorization"));

const keys = h.keys();
for (let i = 0; i < keys.length; i++) {
    console.log(keys[i]);
}

h.delete("content-type");
console.log(h.has("content-type"));
