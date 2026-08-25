// @expect: 0
// @expect: 2
// @expect: 4
// @expect: 6
// @expect: 8
const funcs: (() => number)[] = [];

for (let i = 0; i < 5; i++) {
    funcs.push(() => i * 2);
}

for (const fn of funcs) {
    console.log(fn());
}
