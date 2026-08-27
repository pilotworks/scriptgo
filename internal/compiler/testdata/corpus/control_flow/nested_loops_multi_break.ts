// @expect: 12
// @expect: 50
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

let acc = 0;
for (let x = 0; x < 4; x++) {
    let y = 0;
    while (y < 4) {
        y++;
        if (y === 2) {
            continue;
        }
        acc += (x + y);
    }
}
console.log(acc);
