import { URLSearchParams } from "node:url";
const params = new URLSearchParams("foo=bar");
console.log(params.get("foo"));
