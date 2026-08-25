// @expect: 11
// @expect: 12
// @expect: 99
// @expect: 12
// @expect: 99
function createCounter(initial: number = 0) {
    let count = initial;
    return {
        increment: () => ++count,
        decrement: () => --count,
        getValue: () => count
    };
}

const counter1 = createCounter(10);
const counter2 = createCounter(100);

console.log(counter1.increment());
console.log(counter1.increment());
console.log(counter2.decrement());
console.log(counter1.getValue());
console.log(counter2.getValue());
