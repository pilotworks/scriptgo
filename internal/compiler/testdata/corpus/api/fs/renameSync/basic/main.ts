import * as fs from "node:fs";
fs.writeFileSync("test_old.txt", "rename me");
fs.renameSync("test_old.txt", "test_new.txt");
console.log(fs.existsSync("test_old.txt"));
console.log(fs.existsSync("test_new.txt"));
