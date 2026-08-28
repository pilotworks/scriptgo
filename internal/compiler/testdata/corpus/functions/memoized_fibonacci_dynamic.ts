// @expect: fib(10): 55
// @expect: calls for fib(10): 11
// @expect: fib(15): 610
// @expect: calls for fib(15): 16
function memoize1<T extends number | string, R>(fn: (arg: T) => R): (arg: T) => R {
    const cache = new Map<T, R>();
    return (arg: T): R => {
        if (cache.has(arg)) {
            return cache.get(arg)!;
        }
        const result = fn(arg);
        cache.set(arg, result);
        return result;
    };
}

let fibCalls = 0;
let memoFib: (n: number) => number;

function rawFib(n: number): number {
    fibCalls++;
    if (n <= 0) return 0;
    if (n === 1) return 1;
    return memoFib(n - 1) + memoFib(n - 2);
}

memoFib = memoize1(rawFib);

console.log(`fib(10): ${memoFib(10)}`);
console.log(`calls for fib(10): ${fibCalls}`);

console.log(`fib(15): ${memoFib(15)}`);
console.log(`calls for fib(15): ${fibCalls}`);
