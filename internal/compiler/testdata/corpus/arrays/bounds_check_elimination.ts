// ScriptGo Corpus: Bounds Check Elimination and Loop Vectorization
// @run.expected: 45
// @run.expected: 90
function testBCE(): void {
    const arr: number[] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];
    let sum = 0;
    for (let i = 0; i < arr.length; i++) {
        sum += arr[i];
        arr[i] = arr[i] * 2;
    }
    console.log(sum);

    let sum2 = 0;
    for (let i = 0; i < 10; i++) {
        sum2 += arr[i];
    }
    console.log(sum2);
}

testBCE();
