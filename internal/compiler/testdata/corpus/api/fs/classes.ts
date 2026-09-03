// ScriptGo Corpus: FS Classes & Data Models (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_cls";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });
const sampleFile = tmpDir + "/sample.txt";
fs.writeFileSync(sampleFile, "class models");

// @api: fs.opendirSync
// @expect: true
const dir = fs.opendirSync(tmpDir);
console.log(dir !== null);

// @api: fs.fs.Dir
// @expect: true
console.log(typeof dir === "object");

// @api: fs.Dir.path
// @expect: true
console.log(dir.path.length > 0);

// @api: fs.Dir.readSync
// @expect: true
const entry = dir.readSync();
console.log(entry !== null);

// @api: fs.fs.Dirent
// @expect: true
console.log(typeof entry === "object");

// @api: fs.Dirent.name
// @expect: true
console.log(entry !== null && entry.name.length > 0);

// @api: fs.Dirent.isFile
// @expect: true
console.log(entry !== null && entry.isFile());

// @api: fs.Dirent.isDirectory
// @expect: false
console.log(entry !== null && entry.isDirectory());

// @api: fs.Dirent.isSymbolicLink
// @expect: false
console.log(entry !== null && entry.isSymbolicLink());

// @api: fs.Dirent.isBlockDevice
// @expect: false
console.log(entry !== null && entry.isBlockDevice());

// @api: fs.Dirent.isCharacterDevice
// @expect: false
console.log(entry !== null && entry.isCharacterDevice());

// @api: fs.Dirent.isFIFO
// @expect: false
console.log(entry !== null && entry.isFIFO());

// @api: fs.Dirent.isSocket
// @expect: false
console.log(entry !== null && entry.isSocket());

// @api: fs.Dirent.parentPath
// @expect: true
console.log(entry !== null && entry.parentPath.length >= 0);

// @api: fs.Dirent.path
// @expect: true
console.log(entry !== null && entry.path.length >= 0);

// @api: fs.Dir.closeSync
// @expect: true
dir.closeSync();
console.log(true);

// @api: fs.opendir
// @expect: true
const asyncDir = await fs.promises.opendir(tmpDir);
console.log(asyncDir !== null);

// @api: fs.Dir.read
// @expect: true
const asyncEntry = await asyncDir.read();
console.log(asyncEntry !== null);

// @api: fs.Dir.close
// @expect: true
await asyncDir.close();
console.log(true);

// @api: fs.statSync
// @expect: true
const st = fs.statSync(sampleFile);
console.log(st !== null);

// @api: fs.fs.Stats
// @expect: true
console.log(typeof st === "object");

// @api: fs.Stats.size
// @expect: true
console.log(st.size > 0);

// @api: fs.Stats.isFile
// @expect: true
console.log(st.isFile());

// @api: fs.Stats.isDirectory
// @expect: false
console.log(st.isDirectory());

// @api: fs.Stats.isSymbolicLink
// @expect: false
console.log(st.isSymbolicLink());

// @api: fs.Stats.isBlockDevice
// @expect: false
console.log(st.isBlockDevice());

// @api: fs.Stats.isCharacterDevice
// @expect: false
console.log(st.isCharacterDevice());

// @api: fs.Stats.isFIFO
// @expect: false
console.log(st.isFIFO());

// @api: fs.Stats.isSocket
// @expect: false
console.log(st.isSocket());

// @api: fs.Stats.dev
// @expect: true
console.log(st.dev >= 0);

// @api: fs.Stats.ino
// @expect: true
console.log(st.ino >= 0);

// @api: fs.Stats.mode
// @expect: true
console.log(st.mode > 0);

// @api: fs.Stats.nlink
// @expect: true
console.log(st.nlink >= 1);

// @api: fs.Stats.uid
// @expect: true
console.log(st.uid >= 0);

// @api: fs.Stats.gid
// @expect: true
console.log(st.gid >= 0);

// @api: fs.Stats.rdev
// @expect: true
console.log(st.rdev >= 0);

// @api: fs.Stats.blksize
// @expect: true
console.log(st.blksize >= 0);

// @api: fs.Stats.blocks
// @expect: true
console.log(st.blocks >= 0);

// @api: fs.Stats.atimeMs
// @expect: true
console.log(st.atimeMs > 0);

// @api: fs.Stats.mtimeMs
// @expect: true
console.log(st.mtimeMs > 0);

// @api: fs.Stats.ctimeMs
// @expect: true
console.log(st.ctimeMs > 0);

// @api: fs.Stats.birthtimeMs
// @expect: true
console.log(st.birthtimeMs > 0);

// @api: fs.Stats.atimeNs
// @expect: true
console.log(st.atimeNs > 0);

// @api: fs.Stats.mtimeNs
// @expect: true
console.log(st.mtimeNs > 0);

// @api: fs.Stats.ctimeNs
// @expect: true
console.log(st.ctimeNs > 0);

// @api: fs.Stats.birthtimeNs
// @expect: true
console.log(st.birthtimeNs > 0);

// @api: fs.Stats.atime
// @expect: true
console.log(st.atime.getTime() > 0);

// @api: fs.Stats.mtime
// @expect: true
console.log(st.mtime.getTime() > 0);

// @api: fs.Stats.ctime
// @expect: true
console.log(st.ctime.getTime() > 0);

// @api: fs.Stats.birthtime
// @expect: true
console.log(st.birthtime.getTime() > 0);

// @api: fs.statfsSync
// @expect: true
const sfs = fs.statfsSync(tmpDir);
console.log(sfs !== null);

// @api: fs.fs.StatFs
// @expect: true
console.log(typeof sfs === "object");

// @api: fs.StatFs.bsize
// @expect: true
console.log(sfs.bsize > 0);

// @api: fs.StatFs.blocks
// @expect: true
console.log(sfs.blocks > 0);

// @api: fs.StatFs.bfree
// @expect: true
console.log(sfs.bfree > 0);

// @api: fs.StatFs.bavail
// @expect: true
console.log(sfs.bavail > 0);

// @api: fs.StatFs.files
// @expect: true
console.log(sfs.files > 0);

// @api: fs.StatFs.ffree
// @expect: true
console.log(sfs.ffree > 0);

// Clean up /tmp
fs.rmSync(tmpDir, { recursive: true, force: true });
