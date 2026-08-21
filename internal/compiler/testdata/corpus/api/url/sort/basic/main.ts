import { URLSearchParams } from "node:url";
const params = new URLSearchParams("c=3&a=1&b=2");
params.sort();
console.log(params.toString());
