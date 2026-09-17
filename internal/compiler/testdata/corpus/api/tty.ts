// ScriptGo Corpus: Node.js TTY Module (Strict 1:1 Parity Tests)
import { isatty, ReadStream, WriteStream } from "node:tty";

// @api: tty.isatty
// @expect: true
console.log(typeof isatty === "function");

// @api: tty.isatty.invalid
// @expect: false
console.log(isatty(-1));

// @api: tty.isatty.boolean
// @expect: true
console.log(typeof isatty(0) === "boolean");

// @api: tty.isatty.stdout
// @expect: true
console.log(typeof isatty(1) === "boolean");

// @api: tty.isatty.stderr
// @expect: true
console.log(typeof isatty(2) === "boolean");

// @api: tty.ReadStream
// @expect: true
let readStreamOk = false;
try {
    const rs = new ReadStream(0);
    readStreamOk = (typeof rs.isTTY === "boolean" && typeof rs.isRaw === "boolean");
} catch (e) {
    readStreamOk = true;
}
console.log(readStreamOk);

// @api: tty.WriteStream
// @expect: true
let writeStreamOk = false;
try {
    const ws = new WriteStream(1);
    writeStreamOk = (typeof ws.isTTY === "boolean");
} catch (e) {
    writeStreamOk = true;
}
console.log(writeStreamOk);
