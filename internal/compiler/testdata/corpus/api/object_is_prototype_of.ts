// isPrototypeOf checks the value's class chain against a prototype token.
// @expect: true true false
// @expect: true true false
// @expect: true false
// @expect: false false
class Animal { name = "a"; }
class Dog extends Animal { bark = 1; }
const d = new Dog();
const a = new Animal();
const dogProto = Object.getPrototypeOf(d);
const animalProto = Object.getPrototypeOf(a);
console.log(dogProto.isPrototypeOf(d), animalProto.isPrototypeOf(d), dogProto.isPrototypeOf(a));
const objProto = Object.getPrototypeOf({ x: 1 });
console.log(objProto.isPrototypeOf(d), objProto.isPrototypeOf([1]), objProto.isPrototypeOf(3));
console.log(Object.getPrototypeOf([1]).isPrototypeOf([2, 3]), Object.getPrototypeOf([1]).isPrototypeOf({ y: 2 }));
console.log(({ a: 1 }).isPrototypeOf({ b: 2 }), d.isPrototypeOf(a));
