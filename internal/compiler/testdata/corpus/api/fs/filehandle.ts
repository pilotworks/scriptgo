// ScriptGo Corpus: FileHandle Class & Handle Operations (Strict 1:1 Tests in /tmp)
import * as fs from "node:fs";

const tmpDir = "/tmp/scriptgo_test_fs_fh";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const testFile = tmpDir + "/test_fh.txt";

// @api: fs.promises.open
// @expect: true
const fh = await fs.promises.open(testFile, "w+");
console.log(fh !== null);

// @api: fs.FileHandle
// @expect: true
console.log(typeof fh === "object");

// @api: FileHandle.fd
// @expect: true
console.log(fh.fd >= 0);

// @api: FileHandle.write
// @expect: true
const writeResult = await fh.write("hello filehandle");
console.log(writeResult.bytesWritten > 0);

// @api: FileHandle.sync
// @expect: true
await fh.sync();
console.log(true);

// @api: FileHandle.datasync
// @expect: true
await fh.datasync();
console.log(true);

// @api: FileHandle.stat
// @expect: true
const st = await fh.stat();
console.log(st.isFile());

// @api: FileHandle.chmod
// @expect: true
await fh.chmod(0o644);
console.log(true);

// @api: FileHandle.chown
// @expect: true
await fh.chown(-1, -1);
console.log(true);

// @api: FileHandle.utimes
// @expect: true
await fh.utimes(1600000000, 1600000000);
console.log(true);

// @api: FileHandle.truncate
// @expect: true
await fh.truncate(5);
console.log(true);

// @api: FileHandle.close
// @expect: true
await fh.close();
console.log(true);

// @api: FileHandle.readFile
// @expect: hello
const fhRead = await fs.promises.open(testFile, "r+");
const content = await fhRead.readFile();
console.log(content);

// @api: FileHandle.read
// @expect: true
const readBuf = Buffer.alloc(5);
const nRead = await fhRead.read(readBuf, 0, 5, 0);
console.log(nRead > 0);

// @api: FileHandle.appendFile
// @expect: true
await fhRead.appendFile("!");
console.log(true);

// @api: FileHandle.readv
// @expect: true
const vbuf1 = Buffer.alloc(2);
const vbuf2 = Buffer.alloc(2);
const vread = await fhRead.readv([vbuf1, vbuf2], 0);
console.log(vread >= 0);

// @api: FileHandle.writev
// @expect: true
const vw1 = Buffer.from("a");
const vw2 = Buffer.from("b");
const vwritten = await fhRead.writev([vw1, vw2], 0);
console.log(vwritten >= 0);

await fhRead.close();

// Clean up /tmp
fs.rmSync(tmpDir, { recursive: true, force: true });
