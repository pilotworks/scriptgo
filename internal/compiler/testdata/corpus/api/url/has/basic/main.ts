import { URLSearchParams } from "node:url";
const params = new URLSearchParams("key=val");
console.log(params.has("key"));
console.log(params.has("other"));
