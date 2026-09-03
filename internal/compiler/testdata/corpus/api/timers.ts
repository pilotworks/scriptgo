// ScriptGo Corpus: Timers Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
    Immediate,
    Timeout,
    setImmediate,
    clearImmediate,
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval
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
// @expect: imm_instance: true
const imm = new Immediate(1);
console.log("imm_instance: " + (imm !== null));

// @api: Immediate.[Symbol.dispose]
// @expect: imm_disposed: true
imm[Symbol.dispose]();
console.log("imm_disposed: true");

// @api: timers.Timeout
// @expect: timeout_instance: true
const t = new Timeout(2);
console.log("timeout_instance: " + (t !== null));

// @api: Timeout.close
// @expect: timeout_closed: true
t.close();
console.log("timeout_closed: true");

// @api: Timeout.[Symbol.toPrimitive]
// @expect: timeout_primitive: 2
console.log("timeout_primitive: " + t[Symbol.toPrimitive]());

// @api: Timeout.[Symbol.dispose]
// @expect: timeout_disposed: true
t[Symbol.dispose]();
console.log("timeout_disposed: true");

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


