// @expect: 55
const computed = (() => {
    let temp = 0;
    for (let i = 1; i <= 5; i++) {
        temp += i * i;
    }
    return temp;
})();

console.log(computed);
