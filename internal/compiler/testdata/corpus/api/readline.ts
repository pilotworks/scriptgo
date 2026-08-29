// ScriptGo Corpus: Readline Standard Builtin APIs
// Consolidated test suite with 1:1 isolated assertions for all 30 official Node.js readline APIs.

import {
    InterfaceConstructor,
    Interface,
    createInterface,
    emitKeypressEvents,
    clearLine,
    clearScreenDown,
    cursorTo,
    moveCursor,
    promises
} from "node:readline";

// @api: readline.emitKeypressEvents
// @expect: emitKeypress_ok: true
emitKeypressEvents(null);
console.log("emitKeypress_ok: true");

// @api: new readline.InterfaceConstructor
// @expect: ifc_inst: true
const ifc = new InterfaceConstructor({});
console.log("ifc_inst: " + (ifc instanceof InterfaceConstructor));

// @api: InterfaceConstructor.setPrompt
// @expect: ifc_setPrompt: true
ifc.setPrompt("> ");
console.log("ifc_setPrompt: true");

// @api: InterfaceConstructor.getPrompt
// @expect: ifc_getPrompt: > 
console.log("ifc_getPrompt: " + ifc.getPrompt());

// @api: InterfaceConstructor.prompt
// @expect: ifc_prompt: true
ifc.prompt();
console.log("ifc_prompt: true");

// @api: InterfaceConstructor.write
// @expect: ifc_write: true
ifc.write("hello");
console.log("ifc_write: true");

// @api: InterfaceConstructor.line
// @expect: ifc_line: hello
console.log("ifc_line: " + ifc.line);

// @api: InterfaceConstructor.cursor
// @expect: ifc_cursor: 5
console.log("ifc_cursor: " + ifc.cursor);

// @api: InterfaceConstructor.getCursorPos
// @expect: ifc_cursorPos: 5
console.log("ifc_cursorPos: " + ifc.getCursorPos().cols);

// @api: InterfaceConstructor.pause
// @expect: ifc_pause: true
ifc.pause();
console.log("ifc_pause: true");

// @api: InterfaceConstructor.resume
// @expect: ifc_resume: true
ifc.resume();
console.log("ifc_resume: true");

// @api: InterfaceConstructor.close
// @expect: ifc_close: true
ifc.close();
console.log("ifc_close: true");

// @api: InterfaceConstructor.[Symbol.dispose]
// @expect: ifc_dispose: true
ifc[Symbol.dispose]();
console.log("ifc_dispose: true");

// @api: InterfaceConstructor.[Symbol.asyncIterator]
// @expect: ifc_asyncIter: true
console.log("ifc_asyncIter: " + (ifc[Symbol.asyncIterator]() !== null));

// @api: readline.createInterface
// @expect: createInterface_ok: true
const rl = createInterface({});
console.log("createInterface_ok: " + (rl instanceof Interface));

// @api: readline.Interface
// @expect: rl_inst: true
console.log("rl_inst: " + (rl instanceof Interface));

// @api: readline.Interface.question
// @expect: rl_question: true
rl.question("name? ", (ans: string) => {});
console.log("rl_question: true");

// @api: readline.clearLine
// @expect: clearLine_ok: true
console.log("clearLine_ok: " + clearLine(null, 0));

// @api: readline.clearScreenDown
// @expect: clearScreenDown_ok: true
console.log("clearScreenDown_ok: " + clearScreenDown(null));

// @api: readline.cursorTo
// @expect: cursorTo_ok: true
console.log("cursorTo_ok: " + cursorTo(null, 10));

// @api: readline.moveCursor
// @expect: moveCursor_ok: true
console.log("moveCursor_ok: " + moveCursor(null, 1, 1));

// @api: readline.readlinePromises.Interface
// @expect: pifc_inst: true
const pifc = promises.createInterface({});
console.log("pifc_inst: " + (pifc instanceof promises.Interface));

// @api: readlinePromises.Interface.question
// @expect: pifc_question: true
const pAns = await pifc.question("query?");
console.log("pifc_question: " + (pAns === ""));

// @api: readline.readlinePromises.Readline
// @expect: prl_inst: true
const prl = new promises.Readline(null);
console.log("prl_inst: " + (prl instanceof promises.Readline));

// @api: readlinePromises.Readline.clearLine
// @expect: prl_clearLine: true
console.log("prl_clearLine: " + (prl.clearLine(0) !== null));

// @api: readlinePromises.Readline.clearScreenDown
// @expect: prl_clearScreenDown: true
console.log("prl_clearScreenDown: " + (prl.clearScreenDown() !== null));

// @api: readlinePromises.Readline.cursorTo
// @expect: prl_cursorTo: true
console.log("prl_cursorTo: " + (prl.cursorTo(5) !== null));

// @api: readlinePromises.Readline.moveCursor
// @expect: prl_moveCursor: true
console.log("prl_moveCursor: " + (prl.moveCursor(1, 1) !== null));

// @api: readlinePromises.Readline.rollback
// @expect: prl_rollback: true
console.log("prl_rollback: " + (prl.rollback() !== null));

// @api: readlinePromises.Readline.commit
// @expect: prl_commit: true
await prl.commit();
console.log("prl_commit: true");
