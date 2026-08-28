// @expect: Alice: 50 points
// @expect: 30
// @expect: [true]

function curry2<T, U, R>(f: (a: T, b: U) => R): (a: T) => (b: U) => R {
  return (a: T) => (b: U) => f(a, b);
}

const formatScore = (name: string, score: number): string => `${name}: ${score} points`;
const curriedFormat = curry2(formatScore);
const aliceFormat = curriedFormat("Alice");
console.log(aliceFormat(50));

const multiply = (x: number, y: number): number => x * y;
const curriedMul = curry2(multiply);
const mul10 = curriedMul(10);
console.log(mul10(3));

function wrapInArray<T>(val: T): T[] {
  return [val];
}

const resBool = wrapInArray(true);
console.log(resBool);
