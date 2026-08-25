// @expect: 10
// @expect: 20
// @expect: 15
// @expect: 30
function multiplyBy(factor: number): (x: number) => number {
    return (x: number) => x * factor;
}

const double = multiplyBy(2);
const triple = multiplyBy(3);

console.log(double(5));
console.log(double(10));
console.log(triple(5));
console.log(triple(10));
