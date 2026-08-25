// @expect: 10
// @expect: default
const tuple: [number?, string?] = [];
const [count = 10, label = "default"] = tuple;

console.log(count);
console.log(label);
