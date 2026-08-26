// ScriptGo Corpus Test: AbortController & AbortSignal

import { AbortController, AbortSignal } from "node:events";

// @api: AbortController.constructor
// @expect: false
const controller = new AbortController();
console.log(controller.signal.aborted);

// @api: AbortSignal.addEventListener
// @expect: abort event received
controller.signal.addEventListener("abort", () => {
    console.log("abort event received");
});

// @api: AbortController.abort
// @expect: true
// @expect: cancelled
controller.abort("cancelled");
console.log(controller.signal.aborted);
console.log(controller.signal.reason);

// @api: AbortSignal.throwIfAborted
// @expect: caught: cancelled
try {
    controller.signal.throwIfAborted();
} catch (e) {
    console.log("caught: " + e);
}

// @api: AbortSignal.abort
// @expect: true
// @expect: pre-aborted
const staticSignal = AbortSignal.abort("pre-aborted");
console.log(staticSignal.aborted);
console.log(staticSignal.reason);

// @api: AbortSignal.any
// @expect: true
// @expect: any-aborted
const c1 = new AbortController();
const c2 = new AbortController();
const combined = AbortSignal.any([c1.signal, c2.signal]);
c2.abort("any-aborted");
console.log(combined.aborted);
console.log(combined.reason);

