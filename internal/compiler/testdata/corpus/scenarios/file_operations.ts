// ScriptGo Corpus: Scenario: File Operations
// Consolidated test suite with inline assertions.

import * as fs from "fs";
import { promises } from "node:fs";

// --- Context Case: scenarios_fs_basic ---
// @expect: hello from fs!
// @expect: true
// @expect: false
fs.writeFileSync("test_out.txt", "hello from fs!");
const content_fs_basic_0: string = fs.readFileSync("test_out.txt", "utf-8");
console.log(content_fs_basic_0);
console.log(fs.existsSync("test_out.txt"));
console.log(fs.existsSync("non_existent_file.xyz"));

// --- Context Case: scenarios_fs_extended_sync ---
// @expect: 12
// @expect: true
// @expect: false
// @expect: true
// @expect: Hello World! Appended.
// @expect: true
// @expect: true
// @expect: false
// @expect: 2
// @expect: false
const testDir_fs_extended_sync_1 = "test_fs_extended_tmp";
const mkdirOpts_fs_extended_sync_1: fs.MkdirOptions = { recursive: true };
fs.mkdirSync(testDir_fs_extended_sync_1, mkdirOpts_fs_extended_sync_1);

const filePath_fs_extended_sync_1 = testDir_fs_extended_sync_1 + "/sample.txt";
fs.writeFileSync(filePath_fs_extended_sync_1, "Hello World!");

const stats_fs_extended_sync_1 = fs.statSync(filePath_fs_extended_sync_1);
console.log(stats_fs_extended_sync_1.size);
console.log(stats_fs_extended_sync_1.isFile());
console.log(stats_fs_extended_sync_1.isDirectory());

const dirStats_fs_extended_sync_1 = fs.statSync(testDir_fs_extended_sync_1);
console.log(dirStats_fs_extended_sync_1.isDirectory());

fs.appendFileSync(filePath_fs_extended_sync_1, " Appended.");
console.log(fs.readFileSync(filePath_fs_extended_sync_1));

const copyPath_fs_extended_sync_1 = testDir_fs_extended_sync_1 + "/sample_copy.txt";
fs.copyFileSync(filePath_fs_extended_sync_1, copyPath_fs_extended_sync_1);
console.log(fs.existsSync(copyPath_fs_extended_sync_1));

const renPath_fs_extended_sync_1 = testDir_fs_extended_sync_1 + "/sample_renamed.txt";
fs.renameSync(copyPath_fs_extended_sync_1, renPath_fs_extended_sync_1);
console.log(fs.existsSync(renPath_fs_extended_sync_1));
console.log(fs.existsSync(copyPath_fs_extended_sync_1));

const entries_fs_extended_sync_1 = fs.readdirSync(testDir_fs_extended_sync_1);
console.log(entries_fs_extended_sync_1.length);

fs.rmSync(filePath_fs_extended_sync_1);
fs.rmSync(renPath_fs_extended_sync_1);
const rmOpts_fs_extended_sync_1: fs.RmOptions = { recursive: true, force: true };
fs.rmSync(testDir_fs_extended_sync_1, rmOpts_fs_extended_sync_1);
console.log(fs.existsSync(testDir_fs_extended_sync_1));

// --- Context Case: scenarios_fs_promises ---
// @expect: Promise FS Works!
// @expect: 17
// @expect: true
// @expect: 2
// @expect: false
async function run_fs_promises_2(): Promise<void> {
    const dir_fs_promises_2 = "test_fs_promises_tmp";
    const mkdirOpts_fs_promises_2: fs.MkdirOptions = { recursive: true };
    await promises.mkdir(dir_fs_promises_2, mkdirOpts_fs_promises_2);

    const f_fs_promises_2 = dir_fs_promises_2 + "/data.txt";
    await promises.writeFile(f_fs_promises_2, "Promise FS Works!");
    
    const content_fs_promises_2 = await promises.readFile(f_fs_promises_2);
    console.log(content_fs_promises_2);

    const st_fs_promises_2 = await promises.stat(f_fs_promises_2);
    console.log(st_fs_promises_2.size);
    console.log(st_fs_promises_2.isFile());

    const c_fs_promises_2 = dir_fs_promises_2 + "/data_copy.txt";
    await promises.copyFile(f_fs_promises_2, c_fs_promises_2);

    const entries_fs_promises_2 = await promises.readdir(dir_fs_promises_2);
    console.log(entries_fs_promises_2.length);

    await promises.unlink(f_fs_promises_2);
    await promises.unlink(c_fs_promises_2);
    const rmOpts_fs_promises_2: fs.RmOptions = { recursive: true, force: true };
    await promises.rm(dir_fs_promises_2, rmOpts_fs_promises_2);
    console.log(fs.existsSync(dir_fs_promises_2));
}

run_fs_promises_2();
