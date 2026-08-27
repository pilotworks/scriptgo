// @expect: 10
// @expect: 20
// @expect: 30
// @expect: 60
// @expect: h
// @expect: e
// @expect: y
const nums: number[] = [10, 20, 30];
let total = 0;
for (const n of nums) {
    console.log(n);
    total += n;
}
console.log(total);

const letters: string[] = ["h", "e", "y"];
for (const ch of letters) {
    console.log(ch);
}
