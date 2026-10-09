// @expect: 1 1
// @expect: a
// @expect: 7 3
// Single-element tuples use tuple storage like longer tuples, and an
// element typed undefined is stored boxed.
const one: [number] = [1];
console.log(one[0], one.length);
function first(pair: [string]): string {
  return pair[0];
}
console.log(first(["a"]));
const defaulted = ([x = 7]: [number?] = [undefined]): number => x;
console.log(defaulted(), defaulted([3]));
