// @expect: true true false false
// @expect: false true
// @expect: {"value":1,"writable":true,"enumerable":true,"configurable":true} true
// @expect: 2 true false true
// @expect: object
// Object.getPrototypeOf identity, Object.is identity on objects,
// getOwnPropertyDescriptor and defineProperty on known fields, hasOwn and
// hasOwnProperty, and Object.create(null).
class Base {}
class Derived extends Base {}
const first: unknown = Object.getPrototypeOf({ x: 1 });
const second: unknown = Object.getPrototypeOf({ y: 2 });
const base: unknown = Object.getPrototypeOf(new Base());
const derived: unknown = Object.getPrototypeOf(new Derived());
const list: unknown = Object.getPrototypeOf([1]);
console.log(first === second, base === Object.getPrototypeOf(new Base()), base === derived, list === first);
const left = { a: 1 };
const right = { a: 1 };
console.log(Object.is(left, right), Object.is(left, left));
const source = { a: 1, b: "x" };
const descriptor = Object.getOwnPropertyDescriptor(source, "a");
console.log(JSON.stringify(descriptor), Object.getOwnPropertyDescriptor(source, "zz") === undefined);
const target = { a: 1, b: 0 };
Object.defineProperty(target, "b", { value: 2, writable: true });
console.log(target.b, Object.hasOwn(target, "b"), Object.hasOwn(target, "zz"), target.hasOwnProperty("a"));
const bare: object = Object.create(null);
console.log(typeof bare);
