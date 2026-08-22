// ScriptGo Corpus: Fs Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as fs from "node:fs";

// @api: fs.appendFileSync
// @expect: part1+part2
fs.writeFileSync("test_append.txt", "part1");
fs.appendFileSync("test_append.txt", "+part2");
console.log(fs.readFileSync("test_append.txt", "utf-8"));

// @api: fs.copyFileSync
// @expect: copy me
fs.writeFileSync("test_copy_src.txt", "copy me");
fs.copyFileSync("test_copy_src.txt", "test_copy_dst.txt");
console.log(fs.readFileSync("test_copy_dst.txt", "utf-8"));

// @api: fs.existsSync
// @expect: true
// @expect: false
fs.writeFileSync("test_exists.txt", "content");
console.log(fs.existsSync("test_exists.txt"));
console.log(fs.existsSync("test_non_existent.xyz"));

// @api: fs.mkdirSync
// @expect: true
fs.mkdirSync("test_mkdir");
console.log(fs.existsSync("test_mkdir"));

// @api: fs.readFileSync
// @expect: hello readFileSync
fs.writeFileSync("test_read.txt", "hello readFileSync");
const data_fs_readFileSync_4: string = fs.readFileSync("test_read.txt", "utf-8");
console.log(data_fs_readFileSync_4);

// @api: fs.readdirSync
// @expect: true
fs.mkdirSync("test_dir");
fs.writeFileSync("test_dir/file1.txt", "a");
fs.writeFileSync("test_dir/file2.txt", "b");
const files_fs_readdirSync_5: string[] = fs.readdirSync("test_dir");
console.log(files_fs_readdirSync_5.length === 2);

// @api: fs.renameSync
// @expect: false
// @expect: true
fs.writeFileSync("test_old.txt", "rename me");
fs.renameSync("test_old.txt", "test_new.txt");
console.log(fs.existsSync("test_old.txt"));
console.log(fs.existsSync("test_new.txt"));

// @api: fs.rmSync
// @expect: false
fs.mkdirSync("test_rmdir");
fs.writeFileSync("test_rmdir/test.txt", "hi");
fs.rmSync("test_rmdir", { recursive: true, force: true });
console.log(fs.existsSync("test_rmdir"));

// @api: fs.statSync
// @expect: true
// @expect: true
fs.writeFileSync("test_stat.txt", "12345");
const stats_fs_statSync_8 = fs.statSync("test_stat.txt");
console.log(stats_fs_statSync_8.size > 0);
console.log(stats_fs_statSync_8.isFile());

// @api: fs.unlinkSync
// @expect: true
// @expect: false
fs.writeFileSync("test_unlink.txt", "delete me");
console.log(fs.existsSync("test_unlink.txt"));
fs.unlinkSync("test_unlink.txt");
console.log(fs.existsSync("test_unlink.txt"));

// @api: fs.writeFileSync
// @expect: true
fs.writeFileSync("test_write.txt", "hello writeFileSync");
console.log(fs.existsSync("test_write.txt"));
