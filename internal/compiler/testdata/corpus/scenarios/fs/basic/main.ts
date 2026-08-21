import * as fs from "fs";

fs.writeFileSync("test_out.txt", "hello from fs!");
const content: string = fs.readFileSync("test_out.txt", "utf-8");
console.log(content);
console.log(fs.existsSync("test_out.txt"));
console.log(fs.existsSync("non_existent_file.xyz"));
