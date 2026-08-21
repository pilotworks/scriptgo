import * as fs from "node:fs";
fs.writeFileSync("test_stat.txt", "12345");
const stats = fs.statSync("test_stat.txt");
console.log(stats.size > 0);
console.log(stats.isFile());
