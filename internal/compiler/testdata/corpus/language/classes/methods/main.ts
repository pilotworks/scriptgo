class Calculator {
  value: number = 10;

  add(n: number): number {
    return this.value + n;
  }
}

const c: Calculator = new Calculator();
console.log(c.add(32));
