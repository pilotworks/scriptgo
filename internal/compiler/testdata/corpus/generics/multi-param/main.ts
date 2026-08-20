function makePair<T, U>(a: T, b: U): [T, U] {
    return [a, b];
}

const p = makePair("count", 42);
console.log(p[0]);
console.log(p[1]);
