function factorial(n: number): number {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

function fibonacci(n: number): number {
    if (n <= 0) {
        return 0;
    }
    if (n === 1) {
        return 1;
    }
    return fibonacci(n - 1) + fibonacci(n - 2);
}

console.log(factorial(1));
console.log(factorial(5));
console.log(factorial(7));

console.log(fibonacci(0));
console.log(fibonacci(1));
console.log(fibonacci(6));
console.log(fibonacci(10));
