// @ts-nocheck
// @expect: 1 null 1 s
// @expect: 1 0 a|b 3 1
// @expect: true 1
// new Array(x) with a non-number x, split with an undefined separator or
// limit, and a class field whose initializer always throws.
// TypeScript rejects some of these calls, so type checking is off.
const single = new Array(null);
const word = new Array("s");
console.log(single.length, single[0], word.length, word[0]);
const text = "a b c";
console.log(text.split(undefined).length, text.split(undefined, 0).length, text.split(" ", 2).join("|"), text.split(" ", undefined).length, text.split(" ", "1").length);
function fail() {
  throw new RangeError("init");
}
let built = 0;
class Partial {
  a = ++built;
  b = fail();
  c = ++built;
}
try {
  new Partial();
} catch (e) {
  console.log(e instanceof RangeError, built);
}
