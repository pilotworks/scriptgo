console.log(10);
const p = Promise.resolve(42);
p.then((val) => {
    console.log(val);
});
console.log(20);
