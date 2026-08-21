function greet(name: string, greeting: string = "Hello"): string {
    return greeting + ", " + name + "!";
}

function compute(a: number, b: number = 20, c: number = 5): number {
    return a * b + c;
}

console.log(greet("Alice"));
console.log(greet("Bob", "Good morning"));
console.log(compute(2));
console.log(compute(2, 3));
console.log(compute(2, 3, 4));
