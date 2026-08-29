import { isatty, ReadStream, WriteStream } from "node:tty";

// @api: tty.isatty
// @expect: isatty_0: true
// @expect: isatty_neg: false
console.log("isatty_0: " + isatty(0));
console.log("isatty_neg: " + isatty(-1));

// @api: new tty.tty.ReadStream
// @expect: rs_inst: true
const rs = new ReadStream(0);
console.log("rs_inst: " + (rs instanceof ReadStream));

// @api: tty.ReadStream.isRaw
// @expect: rs_isRaw: false
console.log("rs_isRaw: " + rs.isRaw);

// @api: tty.ReadStream.isTTY
// @expect: rs_isTTY: true
console.log("rs_isTTY: " + rs.isTTY);

// @api: tty.ReadStream.setRawMode
// @expect: rs_setRawMode: true
rs.setRawMode(true);
console.log("rs_setRawMode: " + rs.isRaw);

// @api: new tty.tty.WriteStream
// @expect: ws_inst: true
const ws = new WriteStream(1);
console.log("ws_inst: " + (ws instanceof WriteStream));

// @api: tty.WriteStream.isTTY
// @expect: ws_isTTY: true
console.log("ws_isTTY: " + ws.isTTY);

// @api: tty.WriteStream.columns
// @expect: ws_columns: 80
console.log("ws_columns: " + ws.columns);

// @api: tty.WriteStream.rows
// @expect: ws_rows: 24
console.log("ws_rows: " + ws.rows);

// @api: tty.WriteStream.clearLine
// @expect: ws_clearLine: true
const clRes = ws.clearLine(0);
console.log("ws_clearLine: " + clRes);

// @api: tty.WriteStream.clearScreenDown
// @expect: ws_clearScreenDown: true
const csdRes = ws.clearScreenDown();
console.log("ws_clearScreenDown: " + csdRes);

// @api: tty.WriteStream.cursorTo
// @expect: ws_cursorTo: true
const ctRes = ws.cursorTo(10, 20);
console.log("ws_cursorTo: " + ctRes);

// @api: tty.WriteStream.moveCursor
// @expect: ws_moveCursor: true
const mcRes = ws.moveCursor(2, 3);
console.log("ws_moveCursor: " + mcRes);

// @api: tty.WriteStream.getColorDepth
// @expect: ws_getColorDepth: 24
console.log("ws_getColorDepth: " + ws.getColorDepth());

// @api: tty.WriteStream.hasColors
// @expect: ws_hasColors: true
console.log("ws_hasColors: " + ws.hasColors(256));

// @api: tty.WriteStream.getWindowSize
// @expect: ws_getWindowSize: 80,24
const wsSize = ws.getWindowSize();
console.log("ws_getWindowSize: " + wsSize.join(","));
