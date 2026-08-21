import { EventEmitter } from "node:events";
const ee = new EventEmitter();
ee.once("single", () => {
    console.log("only once");
});
ee.emit("single");
ee.emit("single");
