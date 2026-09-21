// ScriptGo Corpus: node:fs Streams (createReadStream, createWriteStream)
import * as fs from "node:fs";

// @expect: ReadStream opened
// @expect: Chunk: Hello ScriptGo Streaming World!
// @expect: ReadStream closed
// @expect: WriteStream opened
// @expect: WriteStream finished
// @expect: Verified content: Hello ScriptGo Streaming World!
// @expect: FS STREAMS COMPLETED

const tmpDir = "/tmp/scriptgo_test_fs_streams";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const srcFile = tmpDir + "/source.txt";
const dstFile = tmpDir + "/dest.txt";

fs.writeFileSync(srcFile, "Hello ScriptGo Streaming World!");

const rs = fs.createReadStream(srcFile, { encoding: "utf8" });

rs.on("open", (fd: number) => {
    console.log("ReadStream opened");
});

rs.on("data", (chunk: string) => {
    console.log("Chunk:", chunk);
});

rs.on("close", () => {
    console.log("ReadStream closed");

    const rs2 = fs.createReadStream(srcFile);
    const ws = fs.createWriteStream(dstFile);

    ws.on("open", (fd: number) => {
        console.log("WriteStream opened");
    });

    ws.on("finish", () => {
        console.log("WriteStream finished");
        const written = fs.readFileSync(dstFile, "utf8");
        console.log("Verified content:", written);

        fs.rmSync(tmpDir, { recursive: true, force: true });
        console.log("FS STREAMS COMPLETED");
    });

    rs2.pipe(ws);
});
