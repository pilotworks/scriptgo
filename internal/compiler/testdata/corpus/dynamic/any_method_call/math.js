export class Calculator {
  constructor(factor) {
    this.factor = factor;
  }
  scaleAndSum(a, b, c, d, e, f) {
    return (a + b + c + d + e + f) * this.factor;
  }
  greet(name) {
    return `Hello, ${name}!`;
  }
}

export function createCalc(factor) {
  return new Calculator(factor);
}
