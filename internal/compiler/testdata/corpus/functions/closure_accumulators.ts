// @expect: 5
// @expect: 15
// @expect: 25
// @expect: 100
// @expect: 105
function createAccumulator(start: number) {
    let current = start;
    return (delta: number): number => {
        current += delta;
        return current;
    };
}

const acc1 = createAccumulator(0);
console.log(acc1(5));
console.log(acc1(10));
console.log(acc1(10));

const acc2 = createAccumulator(100);
console.log(acc2(0));
console.log(acc2(5));
