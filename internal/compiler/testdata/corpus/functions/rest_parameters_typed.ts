// @expect: Sum: 0
// @expect: Sum: 6
// @expect: Sum: 150
function sumAll(prefix: string, ...values: number[]): string {
    let total = 0;
    for (const v of values) {
        total += v;
    }
    return prefix + ": " + total;
}

console.log(sumAll("Sum"));
console.log(sumAll("Sum", 1, 2, 3));
console.log(sumAll("Sum", 10, 20, 30, 40, 50));
