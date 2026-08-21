console.log("start");

const id1 = setTimeout(() => {
    console.log("should not fire");
}, 10);

const id2 = setTimeout(() => {
    console.log("should fire");
}, 20);

clearTimeout(id1);

console.log("cleared id1");
