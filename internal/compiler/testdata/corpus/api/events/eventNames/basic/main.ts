import { EventEmitter } from "node:events";
const ee = new EventEmitter();
ee.on("foo", () => {});
ee.on("bar", () => {});
const names: string[] = ee.eventNames();
console.log(names.join(","));
