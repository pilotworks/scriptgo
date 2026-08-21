import { EventEmitter } from "node:events";
const ee = new EventEmitter();
ee.on("greet", () => {});
ee.on("greet", () => {});
console.log(ee.listenerCount("greet"));
