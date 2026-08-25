// @expect: 25
function pipe<T>(initial: T, ...ops: ((val: T) => T)[]): T {
    let current = initial;
    for (const op of ops) {
        current = op(current);
    }
    return current;
}

const res = pipe(
    5,
    x => x + 10,
    x => x * 2,
    x => x - 5
);

console.log(res);
