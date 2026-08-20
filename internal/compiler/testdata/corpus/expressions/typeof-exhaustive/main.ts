const num = 42;
const str = "hello";
const boolVal = true;
const bigIntVal = 100n;
const sym = Symbol("test");
const fn = (x: number): number => x * 2;
const arr = [1, 2, 3];
const obj = { a: 1, b: "two" };

console.log(typeof num);
console.log(typeof str);
console.log(typeof boolVal);
console.log(typeof bigIntVal);
console.log(typeof sym);
console.log(typeof fn);
console.log(typeof arr);
console.log(typeof obj);

// Dynamic check on unknown/any
function checkType(val: unknown): string {
  return typeof val;
}

console.log(checkType(123));
console.log(checkType("world"));
console.log(checkType(false));
console.log(checkType(999n));
console.log(checkType(Symbol("dyn")));
