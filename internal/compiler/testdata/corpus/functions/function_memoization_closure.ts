// @expect: computing 5
// @expect: 120
// @expect: cached 5
// @expect: 120
// @expect: computing 6
// @expect: 720
function memoize(fn: (n: number) => number): (n: number) => number {
    const cache = new Map<number, number>();
    return (n: number): number => {
        if (cache.has(n)) {
            console.log("cached " + n);
            return cache.get(n)!;
        }
        console.log("computing " + n);
        const result = fn(n);
        cache.set(n, result);
        return result;
    };
}

function factorial(n: number): number {
    let result = 1;
    for (let i = 2; i <= n; i++) {
        result *= i;
    }
    return result;
}

const memoizedFact = memoize(factorial);
console.log(memoizedFact(5));
console.log(memoizedFact(5));
console.log(memoizedFact(6));
