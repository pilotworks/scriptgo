console.log(1);
queueMicrotask(() => {
    console.log(3);
});
console.log(2);
