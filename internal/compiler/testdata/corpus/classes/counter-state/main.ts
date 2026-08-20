class Counter {
    private count: number;
    constructor(initial: number = 0) {
        this.count = initial;
    }
    increment(): number {
        this.count++;
        return this.count;
    }
    decrement(): number {
        this.count--;
        return this.count;
    }
    getValue(): number {
        return this.count;
    }
}

const c = new Counter(10);
console.log(c.getValue());
console.log(c.increment());
console.log(c.increment());
console.log(c.decrement());
console.log(c.getValue());
