class Stack<T> {
  private items: T[];

  constructor() {
    this.items = [];
  }

  push(item: T): void {
    this.items.push(item);
  }

  pop(): T | undefined {
    if (this.items.length === 0) {
      return undefined;
    }
    return this.items.pop();
  }

  peek(): T | undefined {
    if (this.items.length === 0) {
      return undefined;
    }
    return this.items[this.items.length - 1];
  }

  isEmpty(): boolean {
    return this.items.length === 0;
  }

  size(): number {
    return this.items.length;
  }
}

// Test with numbers
const numStack = new Stack<number>();
console.log(numStack.isEmpty());
numStack.push(10);
numStack.push(20);
numStack.push(30);
console.log(numStack.size());
console.log(numStack.peek());
console.log(numStack.pop());
console.log(numStack.size());
console.log(numStack.isEmpty());

// Test with strings
const strStack = new Stack<string>();
strStack.push("alpha");
strStack.push("beta");
console.log(strStack.peek());
console.log(strStack.pop());
console.log(strStack.pop());
console.log(strStack.pop());
console.log(strStack.isEmpty());
