import { start, builtinModules, REPLServer } from "node:repl";

// @api: repl.builtinModules
// @expect: repl_builtins: true
console.log("repl_builtins: " + (builtinModules.indexOf("fs") !== -1));

// @api: repl.start
// @expect: repl_start_inst: true
const r = start("> ");
console.log("repl_start_inst: " + (r instanceof REPLServer));

// @api: repl.REPLServer
// @expect: repl_server_inst: true
const s = new REPLServer({ prompt: "custom> " });
console.log("repl_server_inst: " + (s instanceof REPLServer));

// @api: REPLServer.defineCommand
// @expect: repl_defineCommand_ok: true
s.defineCommand("test", {
    help: "test command",
    action: () => {}
});
console.log("repl_defineCommand_ok: true");

// @api: REPLServer.displayPrompt
// @expect: repl_displayPrompt_ok: true
s.displayPrompt();
console.log("repl_displayPrompt_ok: true");

// @api: REPLServer.clearBufferedCommand
// @expect: repl_clearBufferedCommand_ok: true
s.clearBufferedCommand();
console.log("repl_clearBufferedCommand_ok: true");

// @api: REPLServer.setupHistory
// @expect: repl_setupHistory_ok: true
s.setupHistory("/tmp/repl_history", () => {});
console.log("repl_setupHistory_ok: true");
