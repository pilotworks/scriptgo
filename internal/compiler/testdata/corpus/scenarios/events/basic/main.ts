import { EventEmitter } from "events";

class UserNotifier extends EventEmitter {
    notify(user: string, count: number): void {
        this.emit("notify", user, count);
    }
}

const emitter = new EventEmitter();

emitter.on("greet", (name: string) => {
    console.log("Hello, " + name);
});

emitter.once("greet", (name: string) => {
    console.log("Once greeting: " + name);
});

emitter.prependListener("greet", (name: string) => {
    console.log("First: " + name);
});

console.log(emitter.listenerCount("greet"));
emitter.emit("greet", "Alice");

console.log(emitter.listenerCount("greet"));
emitter.emit("greet", "Bob");

// Test subclassing
const notifier = new UserNotifier();
notifier.on("notify", (user: string, count: number) => {
    console.log("Notified " + user + " with " + count);
});
notifier.notify("Charlie", 42);

// Test eventNames and removeAllListeners
const names = emitter.eventNames();
console.log(names.length);

emitter.removeAllListeners("greet");
console.log(emitter.listenerCount("greet"));
