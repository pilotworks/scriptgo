import * as fs from "node:fs";
fs.writeFileSync("test_unlink.txt", "delete me");
console.log(fs.existsSync("test_unlink.txt"));
fs.unlinkSync("test_unlink.txt");
console.log(fs.existsSync("test_unlink.txt"));
