import { URLSearchParams } from "node:url";
const params = new URLSearchParams("tag=a&tag=b");
const tags: string[] = params.getAll("tag");
console.log(tags.join(","));
