import { URLSearchParams } from "node:url";
const params = new URLSearchParams("a=1&b=2");
params.delete("a");
console.log(params.toString());
