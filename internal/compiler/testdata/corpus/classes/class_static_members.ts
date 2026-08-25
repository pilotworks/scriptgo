// @expect: 0
// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 0
class Counter {
    static count: number = 0;

    static increment(): number {
        Counter.count++;
        return Counter.count;
    }

    static reset(): void {
        Counter.count = 0;
    }
}

console.log(Counter.count);
console.log(Counter.increment());
console.log(Counter.increment());
console.log(Counter.increment());
Counter.reset();
console.log(Counter.count);
