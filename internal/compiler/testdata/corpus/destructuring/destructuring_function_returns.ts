// @expect: quotient: 3, remainder: 1
// @expect: min: 2, max: 9, avg: 5.5
function divideWithRemainder(a: number, b: number): [number, number] {
    return [Math.floor(a / b), a % b];
}

function computeStats(values: number[]): { min: number; max: number; avg: number } {
    let min = values[0];
    let max = values[0];
    let sum = 0;
    for (let i = 0; i < values.length; i++) {
        if (values[i] < min) min = values[i];
        if (values[i] > max) max = values[i];
        sum += values[i];
    }
    return { min, max, avg: sum / values.length };
}

const [q, r] = divideWithRemainder(10, 3);
console.log("quotient: " + q + ", remainder: " + r);

const { min, max, avg } = computeStats([2, 5, 6, 9]);
console.log("min: " + min + ", max: " + max + ", avg: " + avg);
