import * as fs from "node:fs";
fs.mkdirSync("test_mkdir");
console.log(fs.existsSync("test_mkdir"));
