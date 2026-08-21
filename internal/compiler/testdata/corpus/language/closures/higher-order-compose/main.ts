function makeAdder(x: number): (y: number) => number {
    return (y: number): number => x + y;
}

function applyTwice(f: (n: number) => number, val: number): number {
    return f(f(val));
}

const add5 = makeAdder(5);
console.log(add5(10));
console.log(add5(20));

const add10 = makeAdder(10);
console.log(applyTwice(add10, 5));
