import * as fs from "node:fs";

const testDir = "test_fs_extended_tmp";
const mkdirOpts: fs.MkdirOptions = { recursive: true };
fs.mkdirSync(testDir, mkdirOpts);

const filePath = testDir + "/sample.txt";
fs.writeFileSync(filePath, "Hello World!");

const stats = fs.statSync(filePath);
console.log(stats.size);
console.log(stats.isFile());
console.log(stats.isDirectory());

const dirStats = fs.statSync(testDir);
console.log(dirStats.isDirectory());

fs.appendFileSync(filePath, " Appended.");
console.log(fs.readFileSync(filePath));

const copyPath = testDir + "/sample_copy.txt";
fs.copyFileSync(filePath, copyPath);
console.log(fs.existsSync(copyPath));

const renPath = testDir + "/sample_renamed.txt";
fs.renameSync(copyPath, renPath);
console.log(fs.existsSync(renPath));
console.log(fs.existsSync(copyPath));

const entries = fs.readdirSync(testDir);
console.log(entries.length);

fs.rmSync(filePath);
fs.rmSync(renPath);
const rmOpts: fs.RmOptions = { recursive: true, force: true };
fs.rmSync(testDir, rmOpts);
console.log(fs.existsSync(testDir));
