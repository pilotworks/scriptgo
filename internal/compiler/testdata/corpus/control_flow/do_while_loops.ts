// @expect: 0
// @expect: 1
// @expect: 2
// @expect: executed once: 100
let i = 0;
do {
    console.log(i);
    i++;
} while (i < 3);

let x = 100;
do {
    console.log("executed once: " + x);
} while (x < 10);
