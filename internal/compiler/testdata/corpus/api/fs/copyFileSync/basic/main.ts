import * as fs from "node:fs";
fs.writeFileSync("test_copy_src.txt", "copy me");
fs.copyFileSync("test_copy_src.txt", "test_copy_dst.txt");
console.log(fs.readFileSync("test_copy_dst.txt", "utf-8"));
