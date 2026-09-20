// ScriptGo Corpus: node:fs Watchers (watch, watchFile, unwatchFile)
import * as fs from "node:fs";

// @expect: FS WATCHER INITIALIZED
// @expect: FS WATCH EVENT RECEIVED
// @expect: FS WATCHER CLOSED
// @expect: FS WATCHFILE EVENT RECEIVED
// @expect: FS UNWATCHFILE COMPLETED
// @expect: ALL FS WATCH TESTS PASSED

const tmpDir = "/tmp/scriptgo_test_fs_watch";
if (fs.existsSync(tmpDir)) {
    fs.rmSync(tmpDir, { recursive: true, force: true });
}
fs.mkdirSync(tmpDir, { recursive: true });

const targetFile = tmpDir + "/watched.txt";
fs.writeFileSync(targetFile, "initial content");

console.log("FS WATCHER INITIALIZED");

const watcher = fs.watch(targetFile, (eventType, filename) => {
    console.log("FS WATCH EVENT RECEIVED");
    watcher.close();
});

watcher.on("close", () => {
    console.log("FS WATCHER CLOSED");

    // Test watchFile and unwatchFile
    let called = false;
    const listener = (curr: fs.Stats, prev: fs.Stats) => {
        if (!called) {
            called = true;
            console.log("FS WATCHFILE EVENT RECEIVED");
            fs.unwatchFile(targetFile, listener);
            console.log("FS UNWATCHFILE COMPLETED");
            fs.rmSync(tmpDir, { recursive: true, force: true });
            console.log("ALL FS WATCH TESTS PASSED");
        }
    };

    fs.watchFile(targetFile, { interval: 20 }, listener);

    setTimeout(() => {
        fs.appendFileSync(targetFile, " - modified for watchFile");
    }, 50);
});

setTimeout(() => {
    fs.appendFileSync(targetFile, " - modified for fs.watch");
}, 50);
