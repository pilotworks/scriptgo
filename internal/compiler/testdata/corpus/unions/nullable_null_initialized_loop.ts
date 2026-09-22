// Flow-sensitive union narrowing for null/undefined-initialized variables inside loop
const items = [10, 20];
for (const item of items) {
    let speedup: number | null = null;
    if (item > 15) {
        speedup = 2;
    }
    let res = "none";
    if (speedup !== null) {
        res = (speedup >= 1.0 ? speedup.toFixed(2) + "x" : (1 / speedup).toFixed(2) + "x");
    }
    // @expect: none
    // @expect: 2.00x
    console.log(res);
}
