import { setTimeout, clearTimeout, setInterval, clearInterval, setImmediate } from "node:timers";

console.log("START");
queueMicrotask(() => {
    console.log("MICROTASK");
});

setImmediate(() => {
    console.log("IMMEDIATE");
});

let count = 0;
const timer = setInterval(() => {
    count++;
    console.log("INTERVAL TICK:", count);
    if (count >= 3) {
        clearInterval(timer);
        console.log("CLEARED INTERVAL");
        setTimeout(() => {
            console.log("TIMEOUT DONE");
        }, 5);
    }
}, 5);

