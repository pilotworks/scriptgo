// ScriptGo Corpus: Fs Synchronous Operations & Constants (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_sync";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const testFile = tmpDir + "/test_write.txt";

// @api: fs.writeFileSync
// @expect: true
fs.writeFileSync(testFile, "hello writeFileSync");
console.log(fs.existsSync(testFile));

// @api: fs.readFileSync
// @expect: hello writeFileSync
const readData = fs.readFileSync(testFile, "utf-8");
console.log(readData);

// @api: fs.appendFileSync
// @expect: hello writeFileSync+appended
fs.appendFileSync(testFile, "+appended");
console.log(fs.readFileSync(testFile, "utf-8"));

// @api: fs.copyFileSync
// @expect: true
const copiedFile = tmpDir + "/test_copied.txt";
fs.copyFileSync(testFile, copiedFile);
console.log(fs.existsSync(copiedFile));

// @api: fs.cpSync
// @expect: true
const cpFile = tmpDir + "/test_cp.txt";
fs.cpSync(testFile, cpFile);
console.log(fs.existsSync(cpFile));

// @api: fs.existsSync
// @expect: true
console.log(fs.existsSync(testFile));

// @api: fs.mkdirSync
// @expect: true
const subDir = tmpDir + "/test_mkdir_sub";
fs.mkdirSync(subDir);
console.log(fs.existsSync(subDir));

// @api: fs.readdirSync
// @expect: true
const files = fs.readdirSync(tmpDir);
console.log(files.length > 0);

// @api: fs.renameSync
// @expect: true
const renamedFile = tmpDir + "/test_renamed.txt";
fs.renameSync(copiedFile, renamedFile);
console.log(fs.existsSync(renamedFile));

// @api: fs.rmdirSync
// @expect: false
fs.rmdirSync(subDir);
console.log(fs.existsSync(subDir));

// @api: fs.rmSync
// @expect: false
const toRmFile = tmpDir + "/test_to_rm.txt";
fs.writeFileSync(toRmFile, "to remove");
fs.rmSync(toRmFile, { force: true });
console.log(fs.existsSync(toRmFile));

// @api: fs.statSync
// @expect: true
const st = fs.statSync(testFile);
console.log(st.isFile());

// @api: fs.lstatSync
// @expect: true
const lst = fs.lstatSync(testFile);
console.log(lst.isFile());

// @api: fs.unlinkSync
// @expect: false
fs.unlinkSync(renamedFile);
console.log(fs.existsSync(renamedFile));

// @api: fs.accessSync
// @expect: true
fs.accessSync(testFile, fs.constants.F_OK);
console.log(true);

// @api: fs.constants
// @expect: 0
console.log(fs.constants.F_OK);

// @api: fs.chmodSync
// @expect: true
fs.chmodSync(testFile, 0o644);
console.log(true);

// @api: fs.lchmodSync
// @expect: true
fs.lchmodSync(testFile, 0o644);
console.log(true);

// @api: fs.chownSync
// @expect: true
fs.chownSync(testFile, -1, -1);
console.log(true);

// @api: fs.lchownSync
// @expect: true
fs.lchownSync(testFile, -1, -1);
console.log(true);

// @api: fs.realpathSync
// @expect: true
const realPath = fs.realpathSync(testFile);
console.log(realPath.length > 0);

// @api: fs.truncateSync
// @expect: 5
fs.truncateSync(testFile, 5);
console.log(fs.statSync(testFile).size);

// @api: fs.mkdtempSync
// @expect: true
const tempDir = fs.mkdtempSync(tmpDir + "/tmp_dir_");
console.log(fs.existsSync(tempDir));

// @api: fs.openSync
// @expect: true
const fd = fs.openSync(tmpDir + "/test_open.txt", "w+", 0o666);
console.log(fd >= 0);

// @api: fs.writeSync
// @expect: true
const writtenBytes = fs.writeSync(fd, "hello world");
console.log(writtenBytes > 0);

// @api: fs.readSync
// @expect: true
const readBuf = Buffer.alloc(5);
const readBytes = fs.readSync(fd, readBuf, 0, 5, 0);
console.log(readBytes > 0);

// @api: fs.fstatSync
// @expect: true
const fst = fs.fstatSync(fd);
console.log(fst.size > 0);

// @api: fs.fchmodSync
// @expect: true
fs.fchmodSync(fd, 0o644);
console.log(true);

// @api: fs.fchownSync
// @expect: true
fs.fchownSync(fd, -1, -1);
console.log(true);

// @api: fs.futimesSync
// @expect: true
fs.futimesSync(fd, 1600000000, 1600000000);
console.log(true);

// @api: fs.fsyncSync
// @expect: true
fs.fsyncSync(fd);
console.log(true);

// @api: fs.fdatasyncSync
// @expect: true
fs.fdatasyncSync(fd);
console.log(true);

// @api: fs.ftruncateSync
// @expect: true
fs.ftruncateSync(fd, 3);
console.log(true);

// @api: fs.readvSync
// @expect: true
const vbuf1 = Buffer.alloc(2);
const vbuf2 = Buffer.alloc(2);
const rvSync = fs.readvSync(fd, [vbuf1, vbuf2], 0);
console.log(rvSync >= 0);

// @api: fs.writevSync
// @expect: true
const wvBuf1 = Buffer.from("a");
const wvBuf2 = Buffer.from("b");
const wvSync = fs.writevSync(fd, [wvBuf1, wvBuf2], 0);
console.log(wvSync >= 0);

// @api: fs.closeSync
// @expect: true
fs.closeSync(fd);
console.log(true);

// @api: fs.statfsSync
// @expect: true
const statFsObj = fs.statfsSync(tmpDir);
console.log(statFsObj.bsize > 0);

// @api: fs.linkSync
// @expect: true
const hardLink = tmpDir + "/test_hardlink.txt";
fs.linkSync(testFile, hardLink);
console.log(fs.existsSync(hardLink));

// @api: fs.symlinkSync
// @expect: true
const symLink = tmpDir + "/test_symlink.txt";
fs.symlinkSync(testFile, symLink);
console.log(fs.existsSync(symLink));

// @api: fs.readlinkSync
// @expect: true
const linkTarget = fs.readlinkSync(symLink);
console.log(linkTarget.length > 0);

// @api: fs.utimesSync
// @expect: true
fs.utimesSync(testFile, 1600000000, 1600000000);
console.log(true);

// @api: fs.lutimesSync
// @expect: true
fs.lutimesSync(symLink, 1600000000, 1600000000);
console.log(true);

// @api: fs.globSync
// @expect: true
const matchedFiles = fs.globSync(tmpDir + "/*");
console.log(matchedFiles.length > 0);

// Clean up /tmp
fs.rmSync(tmpDir, { recursive: true, force: true });
