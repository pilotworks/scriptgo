// @expect: 3628800
// @expect: 55
type RecursiveFn<T, R> = (self: (arg: T) => R, arg: T) => R;

function fix<T, R>(fn: RecursiveFn<T, R>): (arg: T) => R {
    return (arg: T): R => fn(fix(fn), arg);
}

const factorial = fix<number, number>((self, n) => {
    if (n <= 1) return 1;
    return n * self(n - 1);
});

console.log(factorial(10));

const fibonacci = fix<number, number>((self, n) => {
    if (n <= 0) return 0;
    if (n === 1) return 1;
    return self(n - 1) + self(n - 2);
});

console.log(fibonacci(10));
