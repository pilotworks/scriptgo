// @expect: 15
// @expect: 26
// @expect: 21
function add(a: number): (b: number) => number {
    return (b: number) => a + b;
}

const add10 = add(10);
console.log(add10(5));

function multiply(factor: number): (val: number) => number {
    return (val: number) => val * factor;
}

const double = multiply(2);
const add6 = add(6);

// compose double then add6: (10 * 2) + 6 = 26
console.log(add6(double(10)));

// (10 + 6) * 2 = 32 -> wait, double(add5(8)) = (8 + 5) * 2 = 26 wait:
const add5 = add(5);
const doubleThenAdd1 = (n: number) => add(1)(double(n));
console.log(doubleThenAdd1(10));
