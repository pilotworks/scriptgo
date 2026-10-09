// @ts-nocheck
// @expect: 4 5
// @expect: false false false true
// @expect: -1 -1 false
// @expect: object true object object
// Static private fields with any name, instanceof on primitives, searches
// for objects in primitive arrays, and Object() conversion.
class Counter {
  static #count = 1;
  static #$ = 2;
  static bump() {
    Counter.#count++;
    return Counter.#count + Counter.#$;
  }
}
console.log(Counter.bump(), Counter.bump());
class K {}
console.log(1 instanceof K, "s" instanceof Object, true instanceof K, new K() instanceof K);
const target = { v: 1 };
console.log([1, 2].indexOf(target), ["a"].lastIndexOf(target), [1].includes([1]));
const list = [1];
const made: object = Object();
console.log(typeof made, Object(list) === list, typeof Object(undefined), typeof Object(null));
