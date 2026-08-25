// @expect: 12
// @expect: 63
let hitCount = 0;

outerLoop: for (let i = 0; i < 5; i++) {
    for (let j = 0; j < 5; j++) {
        if (i === 2 && j === 2) {
            break outerLoop;
        }
        hitCount++;
    }
}

console.log(hitCount);

let sum = 0;
outerFor: for (let i = 1; i <= 3; i++) {
    for (let j = 1; j <= 3; j++) {
        if (j === 2) {
            continue outerFor;
        }
        sum += i * 10 + j;
    }
}

console.log(sum);
