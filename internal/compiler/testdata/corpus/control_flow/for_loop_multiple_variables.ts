// @expect: i=0 j=10 k=100 sum=100
// @expect: i=2 j=9 k=90 sum=208
// @expect: i=4 j=8 k=80 sum=320
// @expect: i=6 j=7 k=70 sum=432
// @expect: Nested count: 85
function multiVarLoop(): void {
    let sum = 0;
    for (let i = 0, j = 10, k = 100; i <= j; i += 2, j--, k -= 10) {
        sum += i * j + k;
        console.log("i=" + i + " j=" + j + " k=" + k + " sum=" + sum);
    }
}

function nestedMutatingLoop(): void {
    let count = 0;
    for (let row = 0, limit = 5; row < limit; row++) {
        for (let col = row; col < 5; col++) {
            if (col === 3) {
                continue;
            }
            if (row === 2 && col === 4) {
                break;
            }
            count += (row + 1) * (col + 1);
        }
    }
    console.log("Nested count: " + count);
}

multiVarLoop();
nestedMutatingLoop();
