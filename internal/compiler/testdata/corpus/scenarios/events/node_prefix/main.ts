import { EventEmitter } from "node:events";

const ee = new EventEmitter();

ee.on("ping", (val: number) => {
    console.log("Ping: " + val);
});

ee.emit("ping", 10);
ee.emit("ping", 25);
console.log(EventEmitter.listenerCount(ee, "ping"));
