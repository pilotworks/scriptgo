// @expect: 10
// @expect: 5
// @expect: 
// @expect: world
// @expect: 100
// @expect: 200
// @expect: 42
let a: number = 0;
let b: number = 5;
let c: string = "";
let d: string = "hello";

a ||= 10;
b ||= 20;
c &&= "world";
d &&= "world";

let x: number | null = null;
let y: number | undefined = undefined;
let z: number = 42;

x ??= 100;
y ??= 200;
z ??= 300;

console.log(a);
console.log(b);
console.log(c);
console.log(d);
console.log(x);
console.log(y);
console.log(z);
