function apply(fn: (x: number) => number, val: number): number {
    return fn(val);
}

const double = (x: number): number => x * 2;
console.log(apply(double, 21));
