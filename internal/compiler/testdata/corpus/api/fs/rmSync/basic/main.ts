import * as fs from "node:fs";
fs.mkdirSync("test_rmdir");
fs.writeFileSync("test_rmdir/test.txt", "hi");
fs.rmSync("test_rmdir", { recursive: true, force: true });
console.log(fs.existsSync("test_rmdir"));
