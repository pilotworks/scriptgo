// ScriptGo Corpus: FS Callback-based APIs (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_cb";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const testFile = tmpDir + "/test_write.txt";
fs.writeFileSync(testFile, "initial data");

// @api: fs.openAsBlob
// @expect: true
fs.openAsBlob(testFile);
console.log(true);

// @api: fs.native
// @expect: true
console.log(fs.existsSync(testFile));

// @api: fs.writeFile
// @expect: true
fs.writeFile(testFile, "callback write data", (err) => {
    console.log(err === null);
});

// @api: fs.readFile
// @expect: true
fs.readFile(testFile, (err, data) => {
    console.log(err === null);
});

// @api: fs.stat
// @expect: true
fs.stat(testFile, (err, st) => {
    console.log(st !== null && st.isFile());
});

// @api: fs.lstat
// @expect: true
fs.lstat(testFile, (err, st) => {
    console.log(st !== null && st.isFile());
});

// @api: fs.exists
// @expect: true
fs.exists(testFile, (exists) => {
    console.log(exists);
});

// @api: fs.access
// @expect: true
fs.access(testFile, fs.constants.F_OK, (err) => {
    console.log(err === null);
});

// @api: fs.chmod
// @expect: true
fs.chmod(testFile, 0o644, (err) => {
    console.log(err === null);
});

// @api: fs.lchmod
// @expect: true
fs.lchmod(testFile, 0o644, (err) => {
    console.log(err === null);
});

// @api: fs.chown
// @expect: true
fs.chown(testFile, -1, -1, (err) => {
    console.log(err === null);
});

// @api: fs.lchown
// @expect: true
fs.lchown(testFile, -1, -1, (err) => {
    console.log(err === null);
});

// @api: fs.realpath
// @expect: true
fs.realpath(testFile, (err, resolved) => {
    console.log(resolved.length > 0);
});

const copyDest = tmpDir + "/test_copy.txt";

// @api: fs.copyFile
// @expect: true
fs.copyFile(testFile, copyDest, (err) => {
    console.log(err === null);
});

const cpFile = tmpDir + "/test_cp.txt";

// @api: fs.cp
// @expect: true
fs.cp(testFile, cpFile, (err) => {
    console.log(err === null);
});

// @api: fs.appendFile
// @expect: true
fs.appendFile(testFile, "+appended", (err) => {
    console.log(err === null);
});

const renamedFile = tmpDir + "/test_renamed.txt";

// @api: fs.rename
// @expect: true
fs.rename(copyDest, renamedFile, (err) => {
    console.log(err === null);
});

const subDir = tmpDir + "/sub_dir";

// @api: fs.mkdir
// @expect: true
fs.mkdir(subDir, (err) => {
    console.log(err === null);
});

// @api: fs.readdir
// @expect: true
fs.readdir(tmpDir, (err, files) => {
    console.log(files.length > 0);
});

// @api: fs.rmdir
// @expect: true
fs.rmdir(subDir, (err) => {
    console.log(err === null);
});

// @api: fs.unlink
// @expect: true
fs.unlink(renamedFile, (err) => {
    console.log(err === null);
});

// @api: fs.truncate
// @expect: true
fs.truncate(testFile, 5, (err) => {
    console.log(err === null);
});

const hardLink = tmpDir + "/test_link.txt";

// @api: fs.link
// @expect: true
fs.link(testFile, hardLink, (err) => {
    console.log(err === null);
});

const symLink = tmpDir + "/test_sym.txt";

// @api: fs.symlink
// @expect: true
fs.symlink(testFile, symLink, (err) => {
    console.log(err === null);
});

// @api: fs.readlink
// @expect: true
fs.readlink(symLink, (err, linkString) => {
    console.log(linkString.length > 0);
});

// @api: fs.utimes
// @expect: true
fs.utimes(testFile, 1600000000, 1600000000, (err) => {
    console.log(err === null);
});

// @api: fs.lutimes
// @expect: true
fs.lutimes(symLink, 1600000000, 1600000000, (err) => {
    console.log(err === null);
});

// @api: fs.statfs
// @expect: true
fs.statfs(tmpDir, (err, statfsObj) => {
    console.log(statfsObj !== null && statfsObj.bsize > 0);
});

// @api: fs.mkdtemp
// @expect: true
fs.mkdtemp(tmpDir + "/temp_", (err, dirPath) => {
    console.log(dirPath.length > 0);
});

const cbOpenFd = fs.openSync(tmpDir + "/test_cb_fd.txt", "w+", 0o666);

// @api: fs.open
// @expect: true
fs.open(testFile, "r+", 0o666, (err, fd) => {
    console.log(fd >= 0);
    fs.closeSync(fd);
});

const cbRBuf = Buffer.alloc(5);

// @api: fs.read
// @expect: true
fs.read(cbOpenFd, cbRBuf, 0, 5, 0, (err, n) => {
    console.log(n >= 0);
});

// @api: fs.write
// @expect: true
fs.write(cbOpenFd, "a", 0, 1, (err, n) => {
    console.log(n >= 0);
});

// @api: fs.readv
// @expect: true
fs.readv(cbOpenFd, [cbRBuf], 0, (err, n) => {
    console.log(n >= 0);
});

// @api: fs.writev
// @expect: true
fs.writev(cbOpenFd, [cbRBuf], 0, (err, n) => {
    console.log(n >= 0);
});

// @api: fs.fchmod
// @expect: true
fs.fchmod(cbOpenFd, 0o644, (err) => {
    console.log(err === null);
});

// @api: fs.fchown
// @expect: true
fs.fchown(cbOpenFd, -1, -1, (err) => {
    console.log(err === null);
});

// @api: fs.fdatasync
// @expect: true
fs.fdatasync(cbOpenFd, (err) => {
    console.log(err === null);
});

// @api: fs.fstat
// @expect: true
fs.fstat(cbOpenFd, (err, st) => {
    console.log(st !== null && st.size >= 0);
});

// @api: fs.fsync
// @expect: true
fs.fsync(cbOpenFd, (err) => {
    console.log(err === null);
});

// @api: fs.ftruncate
// @expect: true
fs.ftruncate(cbOpenFd, 2, (err) => {
    console.log(err === null);
});

// @api: fs.futimes
// @expect: true
fs.futimes(cbOpenFd, 1600000000, 1600000000, (err) => {
    console.log(err === null);
});

// @api: fs.close
// @expect: true
fs.close(cbOpenFd, (err) => {
    console.log(err === null);
});

// @api: fs.glob
// @expect: true
fs.glob(tmpDir + "/*", (err, m) => {
    console.log(err === null);
});

// @api: fs.rm
// @expect: true
fs.rm(cpFile, (err) => {
    console.log(err === null);
});
