// ScriptGo Corpus: Timers Standard Builtin APIs
// Consolidated test suite with inline assertions.

import * as timers from "node:timers";
import {
    setImmediate,
    clearImmediate,
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval,
    promises
} from "node:timers";

// @api: timers.clearImmediate
// @expect: immediate_cleared: true
const immClearId = setImmediate(() => {
    console.log("should not run");
});
clearImmediate(immClearId);
console.log("immediate_cleared: true");

// @api: timers.clearTimeout
// @expect: timeout_cleared: true
const toClearId = setTimeout(() => {
    console.log("should not fire");
}, 10);
clearTimeout(toClearId);
console.log("timeout_cleared: true");

// @api: timers.clearInterval
// @expect: interval_cleared: true
const intClearId = setInterval(() => {
    console.log("should not loop");
}, 20);
clearInterval(intClearId);
console.log("interval_cleared: true");

// @api: timers.Immediate
let imm: timers.Immediate;
if (typeof timers.Immediate === "function") {
    imm = new timers.Immediate(0);
} else {
    imm = setImmediate(() => {}) as unknown as timers.Immediate;
}

// @api: Immediate.hasRef
// @expect: imm_has_ref: true
console.log("imm_has_ref: " + imm.hasRef());

// @api: Immediate.unref
// @expect: imm_unref: true
imm.unref();
console.log("imm_unref: " + (!imm.hasRef()));

// @api: Immediate.ref
// @expect: imm_ref: true
imm.ref();
console.log("imm_ref: " + imm.hasRef());

// @api: Immediate.[Symbol.dispose]
// @expect: imm_disposed: true
imm[Symbol.dispose]();
console.log("imm_disposed: true");

// @api: timers.Timeout
let to: timers.Timeout;
if (typeof timers.Timeout === "function") {
    to = new timers.Timeout(0);
} else {
    to = setTimeout(() => {}, 1000) as unknown as timers.Timeout;
}

// @api: Timeout.hasRef
// @expect: to_has_ref: true
console.log("to_has_ref: " + to.hasRef());

// @api: Timeout.unref
// @expect: to_unref: true
to.unref();
console.log("to_unref: " + (!to.hasRef()));

// @api: Timeout.ref
// @expect: to_ref: true
to.ref();
console.log("to_ref: " + to.hasRef());

// @api: Timeout.refresh
// @expect: to_refresh: true
to.refresh();
console.log("to_refresh: true");

// @api: Timeout.[Symbol.toPrimitive]
// @expect: to_primitive: true
console.log("to_primitive: " + (typeof to[Symbol.toPrimitive]() === "number"));

// @api: Timeout.close
// @expect: to_close: true
to.close();
console.log("to_close: true");

// @api: Timeout.[Symbol.dispose]
// @expect: to_disposed: true
to[Symbol.dispose]();
console.log("to_disposed: true");

// @api: timers.wait
// @expect: scheduler_wait: true
console.log("scheduler_wait: " + (typeof promises.scheduler.wait === "function"));

// @api: timers.yield
// @expect: scheduler_yield: true
console.log("scheduler_yield: " + (typeof promises.scheduler.yield === "function"));

// @api: timers.setImmediate
// @expect: immediate_called: true
setImmediate(() => {
    console.log("immediate_called: true");
});

// @api: timers.setInterval
// @expect: interval_called: true
let intId = 0;
intId = setInterval(() => {
    console.log("interval_called: true");
    clearInterval(intId);
}, 5);

// @api: timers.setTimeout
// @expect: timeout_called: true
setTimeout(() => {
    console.log("timeout_called: true");
}, 10);
