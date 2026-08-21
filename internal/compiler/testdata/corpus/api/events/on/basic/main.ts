import { EventEmitter } from "node:events";
const ee = new EventEmitter();
ee.on("ping", () => {
    console.log("pong");
});
ee.emit("ping");
