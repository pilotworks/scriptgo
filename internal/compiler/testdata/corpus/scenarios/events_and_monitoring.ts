// ScriptGo Corpus: Scenario: Event-Driven & Monitoring
// Consolidated test suite with inline assertions.

import * as consoleMod from "node:console";
import { EventEmitter } from "events";

// --- Context Case: scenarios_console_formatting ---
// @expect: Hello World, count is 42
// @expect: Floating 3.14 and pi
// @expect: Discount 50% off with code SAVE50
// @expect: User Alice has 2 items: item1 item2
console.log("Hello %s, count is %d", "World", 42);
console.log("Floating %f and %s", 3.14, "pi");
console.log("Discount 50%% off with code %s", "SAVE50");
console.log("User %s has %d items:", "Alice", 2, "item1", "item2");

// --- Context Case: scenarios_console_module ---
// @expect: from node:console module
consoleMod.log("from node:console module");

// --- Context Case: scenarios_events_basic ---
// @expect: 3
// @expect: First: Alice
// @expect: Hello, Alice
// @expect: Once greeting: Alice
// @expect: 2
// @expect: First: Bob
// @expect: Hello, Bob
// @expect: Notified Charlie with 42
// @expect: 1
// @expect: 0
class UserNotifier_events_basic_2 extends EventEmitter {
    notify(user: string, count: number): void {
        this.emit("notify", user, count);
    }
}

const emitter_events_basic_2 = new EventEmitter();

emitter_events_basic_2.on("greet", (name: string) => {
    console.log("Hello, " + name);
});

emitter_events_basic_2.once("greet", (name: string) => {
    console.log("Once greeting: " + name);
});

emitter_events_basic_2.prependListener("greet", (name: string) => {
    console.log("First: " + name);
});

console.log(emitter_events_basic_2.listenerCount("greet"));
emitter_events_basic_2.emit("greet", "Alice");

console.log(emitter_events_basic_2.listenerCount("greet"));
emitter_events_basic_2.emit("greet", "Bob");

// Test subclassing
const notifier_events_basic_2 = new UserNotifier_events_basic_2();
notifier_events_basic_2.on("notify", (user: string, count: number) => {
    console.log("Notified " + user + " with " + count);
});
notifier_events_basic_2.notify("Charlie", 42);

// Test eventNames and removeAllListeners
const names_events_basic_2 = emitter_events_basic_2.eventNames();
console.log(names_events_basic_2.length);

emitter_events_basic_2.removeAllListeners("greet");
console.log(emitter_events_basic_2.listenerCount("greet"));

// --- Context Case: scenarios_events_node_prefix ---
// @expect: Ping: 10
// @expect: Ping: 25
// @expect: 1
const ee_events_node_prefix_3 = new EventEmitter();

ee_events_node_prefix_3.on("ping", (val: number) => {
    console.log("Ping: " + val);
});

ee_events_node_prefix_3.emit("ping", 10);
ee_events_node_prefix_3.emit("ping", 25);
console.log(EventEmitter.listenerCount(ee_events_node_prefix_3, "ping"));

// --- Context Case: scenarios_perf_hooks_performance ---
// @expect: true
const t_perf_hooks_performance_4: number = performance.now();
console.log(t_perf_hooks_performance_4 > 0);
