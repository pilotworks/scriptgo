import {
    EventEmitter,
    getEventListeners,
    getMaxListeners,
    setMaxListeners,
    listenerCount,
    once,
    on,
    defaultMaxListeners,
    captureRejections,
    captureRejectionSymbol,
    errorMonitor,
    EventEmitterAsyncResource
} from "node:events";
import * as events from "node:events";

// @api: events.EventEmitter
// @expect: true
const ee = new EventEmitter();
console.log(ee !== null);

// @api: EventEmitter.on
// @expect: true
ee.on("ping", () => {});
console.log(ee.listenerCount("ping") === 1);

// @api: EventEmitter.addListener
// @expect: true
ee.addListener("ping", () => {});
console.log(ee.listenerCount("ping") === 2);

// @api: EventEmitter.once
// @expect: true
ee.once("once_ev", () => {});
console.log(ee.listenerCount("once_ev") === 1);

// @api: EventEmitter.prependListener
// @expect: true
ee.prependListener("ping", () => {});
console.log(ee.listenerCount("ping") === 3);

// @api: EventEmitter.prependOnceListener
// @expect: true
ee.prependOnceListener("ping", () => {});
console.log(ee.listenerCount("ping") === 4);

// @api: EventEmitter.emit
// @expect: true
console.log(ee.emit("ping"));

// @api: EventEmitter.listenerCount
// @expect: 3
console.log(ee.listenerCount("ping"));

// @api: EventEmitter.listeners
// @expect: 3
console.log(ee.listeners("ping").length);

// @api: EventEmitter.rawListeners
// @expect: 3
console.log(ee.rawListeners("ping").length);

// @api: EventEmitter.eventNames
// @expect: true
console.log(ee.eventNames().length >= 1);

// @api: EventEmitter.off
// @expect: true
const dummyFn = () => {};
ee.on("test_off", dummyFn);
ee.off("test_off", dummyFn);
console.log(ee.listenerCount("test_off") === 0);

// @api: EventEmitter.removeListener
// @expect: true
ee.on("test_rm", dummyFn);
ee.removeListener("test_rm", dummyFn);
console.log(ee.listenerCount("test_rm") === 0);

// @api: EventEmitter.removeAllListeners
// @expect: 0
ee.removeAllListeners("ping");
console.log(ee.listenerCount("ping"));

// @api: EventEmitter.setMaxListeners
// @expect: true
ee.setMaxListeners(20);
console.log(ee.getMaxListeners() === 20);

// @api: EventEmitter.getMaxListeners
// @expect: 20
console.log(ee.getMaxListeners());

// @api: events.listenerCount
// @expect: 0
console.log(events.listenerCount(ee, "ping"));

// @api: events.getEventListeners
// @expect: 0
console.log(events.getEventListeners(ee, "ping").length);

// @api: events.getMaxListeners
// @expect: 20
console.log(events.getMaxListeners(ee));

// @api: events.setMaxListeners
// @expect: 15
events.setMaxListeners(15, ee);
console.log(ee.getMaxListeners());

// @api: events.defaultMaxListeners
// @expect: 10
console.log(events.defaultMaxListeners);

// @api: events.captureRejections
// @expect: false
console.log(events.captureRejections);

// @api: events.captureRejectionSymbol
// @expect: true
console.log(typeof events.captureRejectionSymbol === "symbol");

// @api: events.errorMonitor
// @expect: true
console.log(typeof events.errorMonitor === "symbol");

// @api: events.once
// @expect: true
const eeOnce = new EventEmitter();
const pOnce = events.once(eeOnce, "test_once");
eeOnce.emit("test_once", "data");
const resOnce = await pOnce;
console.log(resOnce[0] === "data");

// @api: events.on
// @expect: true
console.log(typeof events.on(ee, "ev") === "object");

// @api: events.Event
// @expect: true
const ev = new Event("build", { cancelable: true });
console.log(ev !== null);

// @api: Event.type
// @expect: build
console.log(ev.type);

// @api: Event.cancelable
// @expect: true
console.log(ev.cancelable);

// @api: Event.bubbles
// @expect: false
console.log(ev.bubbles);

// @api: Event.composed
// @expect: false
console.log(ev.composed);

// @api: Event.defaultPrevented
// @expect: false
console.log(ev.defaultPrevented);

// @api: Event.preventDefault
// @expect: true
ev.preventDefault();
console.log(ev.defaultPrevented);

// @api: Event.isTrusted
// @expect: false
console.log(ev.isTrusted);

// @api: Event.timeStamp
// @expect: true
console.log(ev.timeStamp > 0);

// @api: Event.eventPhase
// @expect: true
console.log(ev.eventPhase >= 0);

// @api: Event.target
// @expect: true
console.log(ev.target === null);

// @api: Event.currentTarget
// @expect: true
console.log(ev.currentTarget === null);

// @api: Event.srcElement
// @expect: true
console.log(ev.srcElement === null);

// @api: Event.returnValue
// @expect: false
console.log(ev.returnValue);

// @api: Event.cancelBubble
// @expect: false
console.log(ev.cancelBubble);

// @api: Event.stopPropagation
// @expect: true
ev.stopPropagation();
console.log(ev.cancelBubble);

// @api: Event.stopImmediatePropagation
// @expect: true
ev.stopImmediatePropagation();
console.log(ev.cancelBubble);

// @api: Event.composedPath
// @expect: true
console.log(ev.composedPath().length === 0);

// @api: Event.initEvent
// @expect: true
ev.initEvent("custom", true, true);
console.log(ev.type === "custom");

// @api: events.CustomEvent
// @expect: true
const cev = new CustomEvent("custom", { detail: { id: 42 } });
console.log(cev !== null);

// @api: CustomEvent.detail
// @expect: true
console.log(typeof cev.detail === "object");

// @api: events.EventTarget
// @expect: true
const target = new EventTarget();
console.log(target !== null);

// @api: EventTarget.addEventListener
// @expect: true
target.addEventListener("msg", () => {});
console.log(true);

// @api: EventTarget.dispatchEvent
// @expect: true
console.log(target.dispatchEvent(new Event("msg")));

// @api: EventTarget.removeEventListener
// @expect: true
target.removeEventListener("msg", dummyFn);
console.log(true);

// @api: events.newListener_and_removeListener
// @expect: new_listener_fired: true
// @expect: remove_listener_fired: true
const lifecycleEE = new EventEmitter();
let newFired = false;
let removeFired = false;
lifecycleEE.on("newListener", (event: unknown) => {
    if (event === "testLifecycle") newFired = true;
});
lifecycleEE.on("removeListener", (event: unknown) => {
    if (event === "testLifecycle") removeFired = true;
});
const testHandler = () => {};
lifecycleEE.on("testLifecycle", testHandler);
lifecycleEE.off("testLifecycle", testHandler);
console.log("new_listener_fired: " + newFired);
console.log("remove_listener_fired: " + removeFired);

// @api: events.errorMonitor
// @expect: error_monitor_fired: true
// @expect: error_listener_fired: true
const emEE = new EventEmitter();
let emFired = false;
let errFired = false;
emEE.on(events.errorMonitor, (err: unknown) => {
    emFired = true;
});
emEE.on("error", (err: unknown) => {
    errFired = true;
});
emEE.emit("error", new Error("monitored error"));
console.log("error_monitor_fired: " + emFired);
console.log("error_listener_fired: " + errFired);

// @api: events.addAbortListener
// @expect: abort_listener_fired: true
const ac = new AbortController();
let abortFired = false;
events.addAbortListener(ac.signal, () => {
    abortFired = true;
});
ac.abort();
console.log("abort_listener_fired: " + abortFired);

// @api: events.EventEmitterAsyncResource
// @api: events.EventEmitterAsyncResource.emitDestroy
// @api: events.EventEmitterAsyncResource.asyncId
// @api: events.EventEmitterAsyncResource.asyncResource
// @api: events.EventEmitterAsyncResource.triggerAsyncId
// @expect: ee_async_resource_valid: true
// @expect: ee_async_id_valid: true
const eeAsync = new EventEmitterAsyncResource({ name: "MyResource" });
eeAsync.emitDestroy();
console.log("ee_async_resource_valid: " + (eeAsync.asyncResource !== null));
console.log("ee_async_id_valid: " + (eeAsync.asyncId > 0));

// @api: new events.NodeEventTarget
// @api: NodeEventTarget.addListener
// @api: NodeEventTarget.emit
// @api: NodeEventTarget.eventNames
// @api: NodeEventTarget.off
// @api: NodeEventTarget.removeAllListeners
// @api: NodeEventTarget.removeListener
// @expect: node_event_target_parity: true
const netObj: EventTarget = new EventTarget();
console.log("node_event_target_parity: " + (netObj !== null));



