import { URLSearchParams } from "node:url";
const params = new URLSearchParams("x=1");
params.set("x", "99");
console.log(params.get("x"));
