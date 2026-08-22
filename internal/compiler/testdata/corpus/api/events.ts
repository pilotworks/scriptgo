// ScriptGo Corpus: Events Standard Builtin APIs
// Consolidated test suite with inline assertions.

import { EventEmitter } from "events";

// @api: events.EventEmitter
// @expect: 1
// @expect: ping: world
const ee_events_EventEmitter_0 = new EventEmitter();
ee_events_EventEmitter_0.on("ping", (name: string) => {
    console.log("ping: " + name);
});
console.log(ee_events_EventEmitter_0.listenerCount("ping"));
ee_events_EventEmitter_0.emit("ping", "world");

// @api: events.emit
// @expect: fired
// @expect: true
// @expect: false
const ee_events_emit_1 = new EventEmitter();
ee_events_emit_1.on("event", () => {
    console.log("fired");
});
console.log(ee_events_emit_1.emit("event"));
console.log(ee_events_emit_1.emit("unregistered"));

// @api: events.eventNames
// @expect: foo,bar
const ee_events_eventNames_2 = new EventEmitter();
ee_events_eventNames_2.on("foo", () => {});
ee_events_eventNames_2.on("bar", () => {});
const names_events_eventNames_2: string[] = ee_events_eventNames_2.eventNames();
console.log(names_events_eventNames_2.join(","));

// @api: events.listenerCount
// @expect: 2
const ee_events_listenerCount_3 = new EventEmitter();
ee_events_listenerCount_3.on("greet", () => {});
ee_events_listenerCount_3.on("greet", () => {});
console.log(ee_events_listenerCount_3.listenerCount("greet"));

// @api: events.off
// @expect: done
const ee_events_off_4 = new EventEmitter();
const cb_events_off_4 = () => {
    console.log("should not fire");
};
ee_events_off_4.on("e", cb_events_off_4);
ee_events_off_4.off("e", cb_events_off_4);
ee_events_off_4.emit("e");
console.log("done");

// @api: events.on
// @expect: pong
const ee_events_on_5 = new EventEmitter();
ee_events_on_5.on("ping", () => {
    console.log("pong");
});
ee_events_on_5.emit("ping");

// @api: events.once
// @expect: only once
const ee_events_once_6 = new EventEmitter();
ee_events_once_6.once("single", () => {
    console.log("only once");
});
ee_events_once_6.emit("single");
ee_events_once_6.emit("single");
