// @expect: 13
// @expect: 22
// @expect: Hello, Ada #2
// @expect: 101,102,103
// bind with partial arguments returns a closure that prepends them; the
// target and the partial arguments are evaluated once, at bind time.
function add3(a: number, b: number, c: number): number { return a + b + c; }
const add10 = add3.bind(null, 10);
console.log(add10(1, 2));
const add15 = add3.bind(null, 10, 5);
console.log(add15(7));
let base = 1;
const greet = (greeting: string, name: string) => greeting + ", " + name + " #" + base;
const hello = greet.bind(undefined, "Hello");
base = 2;
console.log(hello("Ada"));
console.log([1, 2, 3].map(add3.bind(null, 100, 0)).join(","));
