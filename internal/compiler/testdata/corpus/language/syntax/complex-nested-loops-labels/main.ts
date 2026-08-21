let sum: number = 0;

outerLoop: for (let i: number = 0; i < 5; i++) {
    let j: number = 0;
    innerLoop: while (j < 5) {
        j++;
        if (i === 2 && j === 2) {
            console.log("skipping outer at i=2, j=2");
            continue outerLoop;
        }
        if (i === 4 && j === 3) {
            console.log("breaking outer at i=4, j=3");
            break outerLoop;
        }
        if (j === 4) {
            console.log("breaking inner at j=4");
            break innerLoop;
        }
        sum += (i * 10) + j;
    }
}

console.log(sum);
