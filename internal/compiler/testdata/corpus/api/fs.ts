// ScriptGo Corpus: Fs Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as fs from "node:fs";

if (!fs.existsSync("tmp")) {
    fs.mkdirSync("tmp", { recursive: true });
}

// @api: fs.appendFileSync
// @expect: part1+part2
fs.writeFileSync("tmp/test_append.txt", "part1");
fs.appendFileSync("tmp/test_append.txt", "+part2");
console.log(fs.readFileSync("tmp/test_append.txt", "utf-8"));

// @api: fs.copyFileSync
// @expect: copy me
fs.writeFileSync("tmp/test_copy_src.txt", "copy me");
fs.copyFileSync("tmp/test_copy_src.txt", "tmp/test_copy_dst.txt");
console.log(fs.readFileSync("tmp/test_copy_dst.txt", "utf-8"));

// @api: fs.existsSync
// @expect: true
// @expect: false
fs.writeFileSync("tmp/test_exists.txt", "content");
console.log(fs.existsSync("tmp/test_exists.txt"));
console.log(fs.existsSync("tmp/test_non_existent.xyz"));

// @api: fs.mkdirSync
// @expect: true
fs.mkdirSync("tmp/test_mkdir");
console.log(fs.existsSync("tmp/test_mkdir"));

// @api: fs.readFileSync
// @expect: hello readFileSync
fs.writeFileSync("tmp/test_read.txt", "hello readFileSync");
const data_fs_readFileSync_4: string = fs.readFileSync("tmp/test_read.txt", "utf-8");
console.log(data_fs_readFileSync_4);

// @api: fs.readdirSync
// @expect: true
fs.mkdirSync("tmp/test_dir");
fs.writeFileSync("tmp/test_dir/file1.txt", "a");
fs.writeFileSync("tmp/test_dir/file2.txt", "b");
const files_fs_readdirSync_5: string[] = fs.readdirSync("tmp/test_dir");
console.log(files_fs_readdirSync_5.length === 2);

// @api: fs.renameSync
// @expect: false
// @expect: true
fs.writeFileSync("tmp/test_old.txt", "rename me");
fs.renameSync("tmp/test_old.txt", "tmp/test_new.txt");
console.log(fs.existsSync("tmp/test_old.txt"));
console.log(fs.existsSync("tmp/test_new.txt"));

// @api: fs.rmSync
// @expect: false
fs.mkdirSync("tmp/test_rmdir");
fs.writeFileSync("tmp/test_rmdir/test.txt", "hi");
fs.rmSync("tmp/test_rmdir", { recursive: true, force: true });
console.log(fs.existsSync("tmp/test_rmdir"));

// @api: fs.statSync
// @expect: true
// @expect: true
fs.writeFileSync("tmp/test_stat.txt", "12345");
const stats_fs_statSync_8 = fs.statSync("tmp/test_stat.txt");
console.log(stats_fs_statSync_8.size > 0);
console.log(stats_fs_statSync_8.isFile());

// @api: fs.unlinkSync
// @expect: true
// @expect: false
fs.writeFileSync("tmp/test_unlink.txt", "delete me");
console.log(fs.existsSync("tmp/test_unlink.txt"));
fs.unlinkSync("tmp/test_unlink.txt");
console.log(fs.existsSync("tmp/test_unlink.txt"));

// @api: fs.writeFileSync
// @expect: true
fs.writeFileSync("tmp/test_write.txt", "hello writeFileSync");
console.log(fs.existsSync("tmp/test_write.txt"));

// Cleanup test directory
fs.rmSync("tmp", { recursive: true, force: true });

