// ScriptGo Corpus: Node.js Readline Module (Strict 1:1 Parity Tests)
import { EventEmitter } from "node:events";
import {
    Interface,
    createInterface,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    emitKeypressEvents
} from "node:readline";
import {
    createInterface as createInterfacePromises
} from "node:readline/promises";

class MockStream extends EventEmitter {
    resume(): MockStream {
        return this;
    }
    pause(): MockStream {
        return this;
    }
}

// @api: readline.createInterface
// @expect: true
const stream = new MockStream();
const rl = createInterface({ input: stream });
rl.prompt();
console.log(rl.line === "");

// @api: readline.prompt
// @expect: true
rl.setPrompt("test> ");
console.log(rl.getPrompt() === "test> ");

// @api: readline.pause_resume
// @expect: true
rl.pause();
const wasPaused = rl.paused;
rl.resume();
console.log(wasPaused && !rl.paused);

// @api: readline.on_line
// @expect: true
let lineReceived = "";
rl.on("line", (line: string) => {
    lineReceived = line;
});
stream.emit("data", "hello scriptgo\n");
console.log(lineReceived === "hello scriptgo");

// @api: readline.question
// @expect: true
const stream2 = new MockStream();
const rl2 = createInterface({ input: stream2 });
let answerReceived = "";
rl2.question("Enter: ", (ans: string) => {
    answerReceived = ans;
});
stream2.emit("data", "ScriptGo Answer\n");
console.log(answerReceived === "ScriptGo Answer");
rl2.close();

// @api: readline/promises.createInterface
// @expect: true
const stream3 = new MockStream();
const rlp = createInterfacePromises({ input: stream3 });
rlp.prompt();
console.log(rlp.line === "");
rlp.close();

// @api: readline.utilities
// @expect: true
console.log(
    typeof clearLine === "function" &&
    typeof clearScreenDown === "function" &&
    typeof cursorTo === "function" &&
    typeof moveCursor === "function" &&
    typeof emitKeypressEvents === "function"
);
