import * as fs from "node:fs";
fs.writeFileSync("test_read.txt", "hello readFileSync");
const data: string = fs.readFileSync("test_read.txt", "utf-8");
console.log(data);
