function createMultiplier(factor: number): (x: number) => number {
    return (x: number): number => x * factor;
}

function applyTransform(val: number, fn: (n: number) => number): number {
    return fn(val);
}

const double = createMultiplier(2);
const triple = createMultiplier(3);

console.log(double(5));
console.log(triple(5));
console.log(applyTransform(10, double));
console.log(applyTransform(10, triple));

function compose(f: (n: number) => number, g: (n: number) => number): (n: number) => number {
    return (x: number): number => f(g(x));
}

const addTen = (n: number): number => n + 10;
const doubleThenAddTen = compose(addTen, double);
const addTenThenDouble = compose(double, addTen);

console.log(doubleThenAddTen(5));
console.log(addTenThenDouble(5));
