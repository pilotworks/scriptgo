// ScriptGo Corpus: Node.js Readline Module (Strict 1:1 Parity Tests)
import { EventEmitter } from "node:events";
import { PassThrough } from "node:stream";
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
    Interface as PromisesInterface,
    Readline as PromisesReadline,
    createInterface as createInterfacePromises
} from "node:readline/promises";

class MockStream extends EventEmitter {
    write(data: string): boolean {
        return true;
    }
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
console.log(typeof createInterface === "function");

// @api: readline.readline.Interface
// @expect: true
console.log(typeof Interface === "function");

// @api: readline.InterfaceConstructor
// @expect: true
console.log(typeof Interface === "function");

// @api: InterfaceConstructor.setPrompt
rl.setPrompt("test> ");

// @api: InterfaceConstructor.getPrompt
// @expect: true
console.log(rl.getPrompt() === "test> ");

// @api: InterfaceConstructor.prompt
rl.prompt();

// @api: InterfaceConstructor.pause
rl.pause();
const wasPaused = rl.paused;

// @api: InterfaceConstructor.resume
// @expect: true
rl.resume();
console.log(wasPaused && !rl.paused);

// @api: InterfaceConstructor.write
rl.write("input text");

// @api: InterfaceConstructor.getCursorPos
// @expect: true
const pos = rl.getCursorPos();
console.log(typeof pos.rows === "number" && typeof pos.cols === "number");

// @api: InterfaceConstructor.line
// @expect: true
console.log(typeof rl.line === "string");

// @api: InterfaceConstructor.cursor
// @expect: true
console.log(typeof rl.cursor === "number" || typeof rl.cursor === "undefined");

// @api: InterfaceConstructor.[Symbol.asyncIterator]
// @expect: true
const rlIter = rl[Symbol.asyncIterator]();
console.log(rlIter !== null);

// @api: InterfaceConstructor.[Symbol.dispose]
// @api: InterfaceConstructor.close
rl.close();
// @expect: true
console.log(rl.closed);

// @api: readline.Interface.question
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

// @api: readline.emitKeypressEvents
// @expect: true
const kpStream = new EventEmitter();
let kpChar = "";
emitKeypressEvents(kpStream);
kpStream.on("keypress", (ch: string) => {
    kpChar = ch;
});
kpStream.emit("data", "x");
console.log(kpChar === "x");

// @api: readline.clearLine
// @expect: true
console.log(typeof clearLine === "function");

// @api: readline.clearScreenDown
// @expect: true
console.log(typeof clearScreenDown === "function");

// @api: readline.cursorTo
// @expect: true
console.log(typeof cursorTo === "function");

// @api: readline.moveCursor
// @expect: true
console.log(typeof moveCursor === "function");

// @api: readline.readlinePromises.Interface
// @expect: true
console.log(typeof PromisesInterface === "function");

const stream3 = new MockStream();
const rlp: PromisesInterface = createInterfacePromises({ input: stream3 });

// @api: readlinePromises.Interface.question
// @expect: true
const qPromise = rlp.question("Query: ");
stream3.emit("data", "Promise Answer\n");
qPromise.then((ans: string) => {
    console.log(ans === "Promise Answer");
});
rlp.close();

// @api: readline.readlinePromises.Readline
// @expect: true
console.log(typeof PromisesReadline === "function");

const rlpStream = new PassThrough();
const rlpReader = new PromisesReadline(rlpStream);

// @api: readlinePromises.Readline.clearLine
rlpReader.clearLine(1);

// @api: readlinePromises.Readline.clearScreenDown
rlpReader.clearScreenDown();

// @api: readlinePromises.Readline.cursorTo
rlpReader.cursorTo(5, 10);

// @api: readlinePromises.Readline.moveCursor
rlpReader.moveCursor(2, 3);

// @api: readlinePromises.Readline.rollback
rlpReader.rollback();

// @api: readlinePromises.Readline.commit
// @expect: true
rlpReader.cursorTo(1, 2);
rlpReader.commit().then(() => {
    console.log(true);
});
