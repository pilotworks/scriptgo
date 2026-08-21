const add = (a: number, b: number): number => a + b;
const multiply = (a: number, b: number): number => {
    return a * b;
};
const greet = (name: string): string => `Hello, ${name}!`;

console.log(add(10, 20));
console.log(multiply(6, 7));
console.log(greet("TypeScript"));
