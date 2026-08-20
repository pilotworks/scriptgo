class Stack<T> {
  private items: T[] = [];

  push(item: T): void {
    this.items.push(item);
  }

  pop(): T {
    if (this.isEmpty()) {
      throw new Error("StackUnderflow");
    }
    return this.items.pop() as T;
  }

  peek(): T | null {
    return this.items.length > 0 ? this.items[this.items.length - 1] : null;
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }

  size(): number {
    return this.items.length;
  }
}

function evaluateRPN(tokens: string[]): number {
  const stack = new Stack<number>();
  const isNumberRegex: RegExp = /^-?\d+(\.\d+)?$/;

  for (const token of tokens) {
    if (isNumberRegex.test(token)) {
      stack.push(parseFloat(token));
    } else {
      const b = stack.pop();
      const a = stack.pop();
      switch (token) {
        case "+":
          stack.push(a + b);
          break;
        case "-":
          stack.push(a - b);
          break;
        case "*":
          stack.push(a * b);
          break;
        case "/":
          if (b === 0) throw new Error("DivisionByZero");
          stack.push(a / b);
          break;
        default:
          throw new Error(`UnknownOperator: ${token}`);
      }
    }
  }

  return stack.pop();
}

console.log("=== RPN Expression Evaluator ===");
const expr: string[] = ["15", "7", "+", "3", "*", "4", "2", "/", "-"];
try {
  const result = evaluateRPN(expr);
  console.log(`Expression: ${expr.join(" ")}`);
  console.log(`Evaluated Result: ${result}`);
} catch (e) {
  console.log(`Evaluation Error: ${e}`);
}
