// @expect: >> hello
// @expect: ## hello !!
// @expect: 40
// @expect: 42-100
// @expect: 100
// 1. Class method default parameters and optional arguments
class TextFormatter {
    format(content: string, prefix: string = ">>", suffix?: string): string {
        return prefix + " " + content + (suffix ? " " + suffix : "");
    }
}

const formatter: TextFormatter = new TextFormatter();
console.log(formatter.format("hello"));
console.log(formatter.format("hello", "##", "!!"));

// 2. Static method referencing another static method via this
class MathOperations {
    static multiply(n: number): number {
        return n * 2;
    }
    static compute(n: number): number {
        return this.multiply(n) * 2;
    }
}

console.log(MathOperations.compute(10));

// 3. Class static block initialization with numeric array mutations
class Registry {
    static initialCount: number = 42;
    static history: number[] = [];

    static {
        this.history.push(this.initialCount);
        this.initialCount = 100;
        this.history.push(this.initialCount);
    }
}

console.log(Registry.history.join("-"));
console.log(Registry.initialCount);
