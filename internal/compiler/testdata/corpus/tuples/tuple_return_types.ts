// @expect: 5
// @expect: true
// @expect: 0
// @expect: division by zero
function divide(a: number, b: number): [number, string | null] {
    if (b === 0) {
        return [0, "division by zero"];
    }
    return [a / b, null];
}

const [res1, err1] = divide(10, 2);
console.log(res1);
console.log(err1 === null);

const [res2, err2] = divide(10, 0);
console.log(res2);
console.log(err2);
