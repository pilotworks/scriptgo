import * as fs from "node:fs";
fs.writeFileSync("test_append.txt", "part1");
fs.appendFileSync("test_append.txt", "+part2");
console.log(fs.readFileSync("test_append.txt", "utf-8"));
