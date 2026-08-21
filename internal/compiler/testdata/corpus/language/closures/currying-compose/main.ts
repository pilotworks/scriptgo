function add(a: number): (b: number) => number {
  return (b: number): number => a + b;
}

function multiply(a: number): (b: number) => number {
  return (b: number): number => a * b;
}

function compose(f: (x: number) => number, g: (x: number) => number): (x: number) => number {
  return (x: number): number => f(g(x));
}

const add5 = add(5);
const double = multiply(2);
const add5ThenDouble = compose(double, add5);
const doubleThenAdd5 = compose(add5, double);

console.log(add5(10));
console.log(double(10));
console.log(add5ThenDouble(10));
console.log(doubleThenAdd5(10));

// 3-stage pipeline
const square = (x: number): number => x * x;
const pipeline = compose(square, add5ThenDouble);
console.log(pipeline(3));
