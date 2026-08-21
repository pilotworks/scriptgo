import * as fs from "node:fs";
fs.writeFileSync("test_write.txt", "hello writeFileSync");
console.log(fs.existsSync("test_write.txt"));
