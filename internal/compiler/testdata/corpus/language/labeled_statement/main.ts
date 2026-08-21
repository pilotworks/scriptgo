let sum = 0;
outer: for (let i = 0; i < 5; i++) {
    for (let j = 0; j < 5; j++) {
        if (i === 2 && j === 2) {
            break outer;
        }
        sum += 1;
    }
}
console.log(sum);

let contCount = 0;
loopA: for (let i = 0; i < 4; i++) {
    for (let j = 0; j < 4; j++) {
        if (j === 2) {
            continue loopA;
        }
        contCount += 1;
    }
}
console.log(contCount);

let whileSum = 0;
let x = 0;
outerWhile: while (x < 3) {
    x++;
    let y = 0;
    while (y < 3) {
        y++;
        if (x === 2) {
            continue outerWhile;
        }
        whileSum += 1;
    }
}
console.log(whileSum);
