// ScriptGo Corpus: Timers Standard Builtin APIs
// Consolidated test suite with inline assertions.

import {
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

