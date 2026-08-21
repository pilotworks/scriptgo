import * as fs from "node:fs";
fs.writeFileSync("test_exists.txt", "content");
console.log(fs.existsSync("test_exists.txt"));
console.log(fs.existsSync("test_non_existent.xyz"));
