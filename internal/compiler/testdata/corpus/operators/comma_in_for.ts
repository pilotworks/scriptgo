// @expect: 0 10
// @expect: 1 9
// @expect: 2 8
// @expect: done
for (let i = 0, j = 10; i < 3; i++, j--) {
    console.log(i + " " + j);
}
console.log("done");
