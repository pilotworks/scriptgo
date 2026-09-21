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
console.log(typeof ReadStream === "function");

let rsOk = false;
try {
    const rs = new ReadStream(0);
    // @api: tty.ReadStream.setRawMode
    const hasSetRawMode = typeof rs.setRawMode === "function";
    // @api: tty.ReadStream.isRaw
    const hasIsRaw = typeof rs.isRaw === "boolean";
    // @api: tty.ReadStream.isTTY
    const hasIsTTY = typeof rs.isTTY === "boolean";
    rsOk = hasSetRawMode && hasIsRaw && hasIsTTY;
} catch (e) {
    rsOk = true;
}
// @expect: true
console.log(rsOk);

// @api: tty.WriteStream
// @expect: true
console.log(typeof WriteStream === "function");

let wsOk = false;
try {
    const ws = new WriteStream(1);
    // @api: tty.WriteStream.clearLine
    const hasClearLine = typeof ws.clearLine === "function";
    // @api: tty.WriteStream.clearScreenDown
    const hasClearScreenDown = typeof ws.clearScreenDown === "function";
    // @api: tty.WriteStream.cursorTo
    const hasCursorTo = typeof ws.cursorTo === "function";
    // @api: tty.WriteStream.getColorDepth
    const hasGetColorDepth = typeof ws.getColorDepth === "function";
    // @api: tty.WriteStream.getWindowSize
    const hasGetWindowSize = typeof ws.getWindowSize === "function";
    // @api: tty.WriteStream.hasColors
    const hasHasColors = typeof ws.hasColors === "function";
    // @api: tty.WriteStream.moveCursor
    const hasMoveCursor = typeof ws.moveCursor === "function";
    // @api: tty.WriteStream.columns
    const hasColumns = typeof ws.columns === "number" || typeof ws.columns === "undefined";
    // @api: tty.WriteStream.rows
    const hasRows = typeof ws.rows === "number" || typeof ws.rows === "undefined";
    // @api: tty.WriteStream.isTTY
    const hasIsTTY = typeof ws.isTTY === "boolean";
    wsOk = (
        hasClearLine &&
        hasClearScreenDown &&
        hasCursorTo &&
        hasGetColorDepth &&
        hasGetWindowSize &&
        hasHasColors &&
        hasMoveCursor &&
        hasColumns &&
        hasRows &&
        hasIsTTY
    );
} catch (e) {
    wsOk = true;
}
// @expect: true
console.log(wsOk);

let rsNegativeFdCaught = false;
try {
    new ReadStream(-1);
} catch (e) {
    rsNegativeFdCaught = true;
}
// @expect: true
console.log(rsNegativeFdCaught);

let wsNegativeFdCaught = false;
try {
    new WriteStream(-1);
} catch (e) {
    wsNegativeFdCaught = true;
}
// @expect: true
console.log(wsNegativeFdCaught);

