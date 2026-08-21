queueMicrotask(() => {
    console.log("microtask ran");
});
console.log("synchronous log");
