import { URLSearchParams } from "node:url";
const params = new URLSearchParams("a=1");
params.append("b", "2");
console.log(params.toString());
