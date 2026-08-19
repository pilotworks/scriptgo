const a: number[] = [1, 2];
const b: number[] = [3, 4];
const combined: number[] = [...a, 99, ...b];
console.log(combined.length);
console.log(combined[0]);
console.log(combined[1]);
console.log(combined[2]);
console.log(combined[3]);
console.log(combined[4]);

function sumAll(...vals: number[]): number {
    let sum: number = 0;
    for (const val of vals) {
        sum += val;
    }
    return sum;
}

console.log(sumAll(10, 20, 30, 40));
