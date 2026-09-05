// ScriptGo Corpus: FS Callback-based APIs (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_cb";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const testFile = tmpDir + "/test_write.txt";
fs.writeFileSync(testFile, "initial data");
const writeFileTarget = tmpDir + "/write_file.txt";
const appendFileTarget = tmpDir + "/append_file.txt";
const chmodFile = tmpDir + "/chmod_file.txt";
const truncateFile = tmpDir + "/truncate_file.txt";
const utimesFile = tmpDir + "/utimes_file.txt";
fs.writeFileSync(writeFileTarget, "initial");
fs.writeFileSync(appendFileTarget, "initial");
fs.writeFileSync(chmodFile, "initial");
fs.writeFileSync(truncateFile, "initial data");
fs.writeFileSync(utimesFile, "initial");

// @api: fs.writeFile
// @expect: true
fs.writeFile(writeFileTarget, "callback write data", (err) => {
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
fs.chmod(chmodFile, 0o644, (err) => {
    console.log(err === null);
});

const chownFile = tmpDir + "/test_chown.txt";
const lchownFile = tmpDir + "/test_lchown.txt";
fs.writeFileSync(chownFile, "chown");
fs.writeFileSync(lchownFile, "lchown");

// @api: fs.chown
// @expect: true
fs.chown(chownFile, -1, -1, (err) => {
    console.log(err === null);
});

// @api: fs.lchown
// @expect: true
fs.lchown(lchownFile, -1, -1, (err) => {
    console.log(err === null);
});

// @api: fs.realpath
// @expect: true
fs.realpath(testFile, (err, resolved) => {
    console.log(resolved.length > 0);
});

const copyDest = tmpDir + "/test_copy.txt";
const renamedFile = tmpDir + "/test_renamed.txt";
fs.writeFileSync(renamedFile, "rename target");

// @api: fs.copyFile
// @expect: true
fs.copyFile(testFile, copyDest, (err) => {
    console.log(err === null);
    // @expect: true
    fs.rename(copyDest, renamedFile, (renameErr) => {
        console.log(renameErr === null);
    });
});

const cpFile = tmpDir + "/test_cp.txt";
fs.writeFileSync(cpFile, "cp target");

// @api: fs.cp
// @expect: true
fs.cp(testFile, cpFile, (err) => {
    console.log(err === null);
});

// @api: fs.appendFile
// @expect: true
fs.appendFile(appendFileTarget, "+appended", (err) => {
    console.log(err === null);
});

const subDir = tmpDir + "/sub_dir";
const rmdirDir = tmpDir + "/rmdir_dir";
fs.mkdirSync(rmdirDir);

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
fs.rmdir(rmdirDir, (err) => {
    console.log(err === null);
});

// @api: fs.unlink
// @expect: true
const unlinkFile = tmpDir + "/test_unlink.txt";
fs.writeFileSync(unlinkFile, "unlink");
fs.unlink(unlinkFile, (err) => {
    console.log(err === null);
});

// @api: fs.truncate
// @expect: true
fs.truncate(truncateFile, 5, (err) => {
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
// @api: fs.readlink
// @expect: true
fs.symlink(testFile, symLink, (err) => {
    if (err !== null) {
        console.log(false);
        return;
    }
    fs.readlink(symLink, (readErr, linkString) => {
        console.log(readErr === null && linkString !== undefined && linkString.length > 0);
    });
});

// @api: fs.utimes
// @expect: true
fs.utimes(utimesFile, 1600000000, 1600000000, (err) => {
    console.log(err === null);
});

// @api: fs.lutimes
// @expect: true
fs.lutimes(utimesFile, 1600000000, 1600000000, (err) => {
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
const closeOnlyFd = fs.openSync(tmpDir + "/test_close_fd.txt", "w");
fs.close(closeOnlyFd, (err) => {
    console.log(err === null);
});

// @api: fs.rm
// @expect: true
const rmFile = tmpDir + "/test_rm.txt";
fs.writeFileSync(rmFile, "rm");
fs.rm(rmFile, (err) => {
    console.log(err === null);
});
