const id = setImmediate(() => {
    console.log("should not run");
});
clearImmediate(id);
console.log("immediate cleared");
