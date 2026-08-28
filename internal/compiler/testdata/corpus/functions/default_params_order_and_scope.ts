// @expect: 30
// @expect: 16
// @expect: 115

function compute(a: number, b: number = a * 2, c: number = a + b): number {
  return a + b + c;
}

console.log(compute(5));
console.log(compute(5, 3));
console.log(compute(5, undefined, 100));
