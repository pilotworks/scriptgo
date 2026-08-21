import * as fs from "node:fs";
fs.mkdirSync("test_dir");
fs.writeFileSync("test_dir/file1.txt", "a");
fs.writeFileSync("test_dir/file2.txt", "b");
const files: string[] = fs.readdirSync("test_dir");
console.log(files.length === 2);
