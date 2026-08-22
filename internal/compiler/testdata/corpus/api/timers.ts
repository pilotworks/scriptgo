// ScriptGo Corpus: Timers Standard Builtin APIs
// Consolidated test suite with inline assertions.

// @api: timers.clearImmediate
// @expect: immediate cleared
const id_timers_clearImmediate_0 = setImmediate(() => {
    console.log("should not run");
});
clearImmediate(id_timers_clearImmediate_0);
console.log("immediate cleared");

// @api: timers.clearTimeout
// @expect: start
// @expect: cleared id1
console.log("start");

const id1_timers_clearTimeout_1 = setTimeout(() => {
    console.log("should not fire");
}, 10);

const id2_timers_clearTimeout_1 = setTimeout(() => {
    console.log("should fire");
}, 20);

clearTimeout(id1_timers_clearTimeout_1);

console.log("cleared id1");

// @api: timers.setImmediate
// @expect: before immediate
setImmediate(() => {
    console.log("immediate ran");
});
console.log("before immediate");

// @api: timers.setInterval
// @expect: start
console.log("start");

let timerId_timers_setInterval_3 = 0;
timerId_timers_setInterval_3 = setInterval(() => {
    console.log("interval fired");
    clearInterval(timerId_timers_setInterval_3);
    console.log("cleared interval");
}, 10);

// @api: timers.setTimeout
// @expect: start
// @expect: end
// @expect: immediate ran
// @expect: timeout 0ms
// @expect: interval fired
// @expect: cleared interval
// @expect: timeout 10ms
// @expect: should fire
console.log("start");

setTimeout(() => {
    console.log("timeout 10ms");
}, 10);

setTimeout(() => {
    console.log("timeout 0ms");
}, 0);

console.log("end");
