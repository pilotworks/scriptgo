import { EventEmitter } from "events";

const ee = new EventEmitter();
ee.on("ping", (name: string) => {
    console.log("ping: " + name);
});
console.log(ee.listenerCount("ping"));
ee.emit("ping", "world");
