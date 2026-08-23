// @expect: 1
// @expect: 2
// @expect: 3
// @expect: 4
// @expect: 5
const data = {
    items: [1, 2, 3],
    nested: {
        points: [4, 5],
    },
};

const { items: [first, second, third], nested: { points: [p1, p2] } } = data;
console.log(first);
console.log(second);
console.log(third);
console.log(p1);
console.log(p2);
