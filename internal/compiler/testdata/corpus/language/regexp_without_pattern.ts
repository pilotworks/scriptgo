// @ts-nocheck
// @expect: (?:) true -abc
// @expect: (?:) true
// RegExp() and new RegExp() without a pattern match the empty string.
const empty = new RegExp();
console.log(empty.source, empty.test("anything"), "abc".replace(empty, "-"));
const called = RegExp();
console.log(called.source, called.test(""));
