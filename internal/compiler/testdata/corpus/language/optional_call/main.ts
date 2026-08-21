function greet(name: string): string {
    return "Hello, " + name;
}

class Greeter {
    greeting: string;
    constructor(greeting: string) {
        this.greeting = greeting;
    }
    say(name: string): string {
        return this.greeting + ", " + name + "!";
    }
}

console.log(greet?.("Alice"));

const greeter = new Greeter("Welcome");
console.log(greeter?.say?.("Bob"));

const fn = (x: number) => x * 2;
console.log(fn?.(21));
