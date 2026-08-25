// @expect: 42
// @expect: 999
let sideEffect = 0;

function compute(): number {
    try {
        return 42;
    } finally {
        sideEffect = 999;
    }
}

const result = compute();
console.log(result);
console.log(sideEffect);
