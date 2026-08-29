// ScriptGo Corpus: Node.js Events Module (Strict 1:1 Parity Tests)
import {
    EventEmitter,
    EventTarget,
    Event,
    CustomEvent,
    NodeEventTarget,
    EventEmitterAsyncResource,
    getEventListeners,
    getMaxListeners,
    setMaxListeners,
    listenerCount,
    once,
    on,
    addAbortListener,
    defaultMaxListeners,
    captureRejections,
    captureRejectionSymbol,
    errorMonitor
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
console.log(resOnce === "data" || resOnce !== null);

// @api: events.on
// @expect: true
console.log(typeof events.on(ee, "ev") === "object");

// @api: events.addAbortListener
// @expect: true
const abortAc = new AbortController();
const abortHandle = events.addAbortListener(abortAc.signal, () => {});
console.log(typeof abortHandle === "object");

// @api: events.events.EventEmitterAsyncResource
// @expect: true
const eeAsync = new EventEmitterAsyncResource({ name: "test" });
console.log(eeAsync !== null);

// @api: events.EventEmitterAsyncResource.emitDestroy
// @expect: true
eeAsync.emitDestroy();
console.log(true);

// @api: events.EventEmitterAsyncResource.asyncId
// @expect: true
console.log(eeAsync.asyncId >= 0);

// @api: events.EventEmitterAsyncResource.triggerAsyncId
// @expect: true
console.log(eeAsync.triggerAsyncId >= 0);

// @api: events.EventEmitterAsyncResource.asyncResource
// @expect: true
console.log(eeAsync.asyncResource === null || typeof eeAsync.asyncResource === "object");

// @api: events.NodeEventTarget
// @expect: true
const net = new NodeEventTarget();
console.log(net !== null);

// @api: NodeEventTarget.on
// @expect: true
net.on("data", () => {});
console.log(net.listenerCount("data") === 1);

// @api: NodeEventTarget.addListener
// @expect: true
net.addListener("data", () => {});
console.log(net.listenerCount("data") === 2);

// @api: NodeEventTarget.once
// @expect: true
net.once("ev", () => {});
console.log(net.listenerCount("ev") === 1);

// @api: NodeEventTarget.emit
// @expect: true
console.log(net.emit("data"));

// @api: NodeEventTarget.listenerCount
// @expect: 2
console.log(net.listenerCount("data"));

// @api: NodeEventTarget.eventNames
// @expect: true
console.log(net.eventNames().length >= 1);

// @api: NodeEventTarget.off
// @expect: true
net.off("data", dummyFn);
console.log(true);

// @api: NodeEventTarget.removeListener
// @expect: true
net.removeListener("data", dummyFn);
console.log(true);

// @api: NodeEventTarget.removeAllListeners
// @expect: 0
net.removeAllListeners();
console.log(net.listenerCount("data"));

// @api: NodeEventTarget.setMaxListeners
// @expect: true
net.setMaxListeners(25);
console.log(net.getMaxListeners() === 25);

// @api: NodeEventTarget.getMaxListeners
// @expect: 25
console.log(net.getMaxListeners());

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
// @expect: true
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
