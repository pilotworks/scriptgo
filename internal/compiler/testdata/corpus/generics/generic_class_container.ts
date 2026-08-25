// @expect: 3
// @expect: 30
// @expect: 30
// @expect: 20
// @expect: 1
// @expect: world
class Stack<T> {
    private items: T[] = [];

    push(item: T): void {
        this.items.push(item);
    }

    pop(): T | undefined {
        return this.items.pop();
    }

    peek(): T | undefined {
        return this.items.length > 0 ? this.items[this.items.length - 1] : undefined;
    }

    size(): number {
        return this.items.length;
    }
}

const numStack = new Stack<number>();
numStack.push(10);
numStack.push(20);
numStack.push(30);

console.log(numStack.size());
console.log(numStack.peek());
console.log(numStack.pop());
console.log(numStack.pop());
console.log(numStack.size());

const strStack = new Stack<string>();
strStack.push("hello");
strStack.push("world");
console.log(strStack.pop());
