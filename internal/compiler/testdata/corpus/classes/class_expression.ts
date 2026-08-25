// @expect: Hello, TypeScript!
const Greeter = class {
    constructor(public greeting: string) {}
    greet(name: string): string {
        return this.greeting + ", " + name + "!";
    }
};

const g = new Greeter("Hello");
console.log(g.greet("TypeScript"));
