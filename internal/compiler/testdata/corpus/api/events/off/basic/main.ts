import { EventEmitter } from "node:events";
const ee = new EventEmitter();
const cb = () => {
    console.log("should not fire");
};
ee.on("e", cb);
ee.off("e", cb);
ee.emit("e");
console.log("done");
