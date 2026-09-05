// ScriptGo Corpus: Fs Promises APIs (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_promises";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const testFile = tmpDir + "/test_promise.txt";

// @api: fs.promises.writeFile
// @expect: true
await fs.promises.writeFile(testFile, "promise content");
console.log(fs.existsSync(testFile));

// @api: fs.promises.readFile
// @expect: <Buffer 70 72 6f 6d 69 73 65 20 63 6f 6e 74 65 6e 74>
const pData = await fs.promises.readFile(testFile);
console.log(pData);

// @api: fs.promises.appendFile
// @expect: <Buffer 70 72 6f 6d 69 73 65 20 63 6f 6e 74 65 6e 74 2b 61 70 70 65 6e 64 65 64>
await fs.promises.appendFile(testFile, "+appended");
const pAppendData = await fs.promises.readFile(testFile);
console.log(pAppendData);

// @api: fs.promises.copyFile
// @expect: true
const copyDest = tmpDir + "/test_copied.txt";
await fs.promises.copyFile(testFile, copyDest);
console.log(fs.existsSync(copyDest));

// @api: fs.promises.cp
// @expect: true
const cpFile = tmpDir + "/test_cp.txt";
await fs.promises.cp(testFile, cpFile);
console.log(fs.existsSync(cpFile));

// @api: fs.promises.rename
// @expect: true
const renamedFile = tmpDir + "/test_renamed.txt";
await fs.promises.rename(copyDest, renamedFile);
console.log(fs.existsSync(renamedFile));

// @api: fs.promises.stat
// @expect: true
const pStat = await fs.promises.stat(testFile);
console.log(pStat.isFile());

// @api: fs.promises.lstat
// @expect: true
const pLStat = await fs.promises.lstat(testFile);
console.log(pLStat.isFile());

// @api: fs.promises.mkdir
// @expect: true
const nestedDir = tmpDir + "/nested_dir";
await fs.promises.mkdir(nestedDir, { recursive: true });
console.log(fs.existsSync(nestedDir));

// @api: fs.promises.readdir
// @expect: true
const pFiles = await fs.promises.readdir(tmpDir);
console.log(pFiles.length > 0);

// @api: fs.promises.access
// @expect: true
await fs.promises.access(testFile);
console.log(true);

// @api: fs.promises.chmod
// @expect: true
await fs.promises.chmod(testFile, 0o644);
console.log(true);

// @api: fs.promises.realpath
// @expect: true
const pReal = await fs.promises.realpath(testFile);
console.log(pReal.length > 0);

// @api: fs.promises.truncate
// @expect: 7
await fs.promises.truncate(testFile, 7);
const pTruncStat = await fs.promises.stat(testFile);
console.log(pTruncStat.size);

// @api: fs.promises.mkdtemp
// @expect: true
const pTemp = await fs.promises.mkdtemp(tmpDir + "/temp_");
console.log(fs.existsSync(pTemp));

// @api: fs.promises.statfs
// @expect: true
const pStatfs = await fs.promises.statfs(tmpDir);
console.log(pStatfs.bsize > 0);

// @api: fs.promises.link
// @expect: true
const hardLink = tmpDir + "/hardlink.txt";
await fs.promises.link(testFile, hardLink);
console.log(fs.existsSync(hardLink));

// @api: fs.promises.symlink
// @expect: true
const symLink = tmpDir + "/symlink.txt";
await fs.promises.symlink(testFile, symLink);
console.log(fs.existsSync(symLink));

// @api: fs.promises.readlink
// @expect: true
const pReadlink = await fs.promises.readlink(symLink);
console.log(pReadlink.length > 0);

// @api: fs.promises.utimes
// @expect: true
await fs.promises.utimes(testFile, 1600000000, 1600000000);
console.log(true);

// @api: fs.promises.lutimes
// @expect: true
await fs.promises.lutimes(symLink, 1600000000, 1600000000);
console.log(true);

// @api: fs.promises.rm
// @expect: false
await fs.promises.rm(renamedFile);
console.log(fs.existsSync(renamedFile));

// @api: fs.promises.rmdir
// @expect: false
await fs.promises.rmdir(nestedDir);
console.log(fs.existsSync(nestedDir));

// @api: fs.promises.unlink
// @expect: false
await fs.promises.unlink(testFile);
console.log(fs.existsSync(testFile));

// Clean up /tmp
fs.rmSync(tmpDir, { recursive: true, force: true });
