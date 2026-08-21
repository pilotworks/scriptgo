console.log("start");

setTimeout(() => {
    console.log("timeout 10ms");
}, 10);

setTimeout(() => {
    console.log("timeout 0ms");
}, 0);

console.log("end");
