// Functions read top-level objects and namespace members at run time, so
// later mutations and reassignments are visible.
import * as fs from "node:fs";
import * as os from "node:os";
import { EventEmitter } from "node:events";

// @expect: 2
const state = { count: 0 };
function inc() { state.count++; }
function get() { return state.count; }
inc();
inc();
console.log(get());

// @expect: 5
const cfg = { db: { host: "h", port: 5 } };
function port() { return cfg.db.port; }
console.log(port());

// @expect: 9
let o = { v: 1 };
function readV() { return o.v; }
o = { v: 9 };
console.log(readV());

// @expect: 0
function access() { return fs.constants.F_OK; }
console.log(access());

// @expect: 20
EventEmitter.defaultMaxListeners = 20;
function limit() { return EventEmitter.defaultMaxListeners; }
console.log(limit());

// Same-named module variables stay separate: this module's `constants` is
// not fs's or os's.
// @expect: 7 2
const constants = { mine: 7 };
function mine() { return constants.mine; }
function sigint() { return os.constants.signals.SIGINT; }
console.log(mine(), sigint());
