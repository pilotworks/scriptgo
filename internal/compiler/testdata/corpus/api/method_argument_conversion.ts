// @ts-nocheck
// @expect: b bcdef 3 1,2
// @expect: a cdef b 2
// @expect: 0 12 NaN 16 7 NaN 0
// @expect: 000ab 1 true abc  |
// @expect: 3 -1 false 1 0 true
// @expect: map true
// @expect: 1,2,3
// @expect: symbol true
// Built-in string/array methods convert their arguments as the spec does:
// ToNumber/ToIntegerOrInfinity for indices (undefined is 0, or "not present"
// for a trailing end), ToString for search and fill strings, strict
// equality for searches, and a TypeError for a non-callable callback.
// TypeScript rejects these calls, so this file opts out of type checking.
console.log("abc".charAt(true), "abcdef".slice("1", undefined), "aXbX".indexOf("X", "2"), [1, 2, 3].slice(false, "2").join());
console.log("abc".charAt(undefined), "abcdef".substring(2, undefined), "abc".charAt([1]), [1, 2, 3].indexOf(3, [1]));
console.log(Number(""), Number(" 12 "), Number("12px"), Number("0x10"), Number([7]), Number([1, 2]), Number(null));
console.log("ab".padStart(5, 0), "a1b".indexOf(1), "xtrue".endsWith(true), "abc".padEnd(5, undefined) + "|");
const nums = [1, 2, 3, 1];
const flags = [true, false, true];
console.log(nums.indexOf(1, -1), nums.indexOf(true), nums.includes("1"), flags.indexOf(false), flags.lastIndexOf(true, 1), flags.includes(false));
try {
  [1, 2].map(true);
} catch (e) {
  console.log("map", e instanceof TypeError);
}
console.log([3, 1, 2].sort(undefined).join());
try {
  "abc".charAt(Symbol("s"));
} catch (e) {
  console.log("symbol", e instanceof TypeError);
}
