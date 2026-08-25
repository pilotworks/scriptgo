// @expect: age
// @expect: 30
// @expect: 30
// @expect: age
function makePair<T, U>(first: T, second: U): [T, U] {
    return [first, second];
}

function swapPair<T, U>(pair: [T, U]): [U, T] {
    return [pair[1], pair[0]];
}

const p1 = makePair<string, number>("age", 30);
console.log(p1[0]);
console.log(p1[1]);

const p2 = swapPair(p1);
console.log(p2[0]);
console.log(p2[1]);
