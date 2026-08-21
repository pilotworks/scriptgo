import { EventEmitter } from "node:events";
const ee = new EventEmitter();
ee.on("event", () => {
    console.log("fired");
});
console.log(ee.emit("event"));
console.log(ee.emit("unregistered"));
