// String.prototype.repeat follows ToIntegerOrInfinity and throws RangeError
// for negative or infinite counts and over-long results; repeating "" is
// immediate even for a huge count.
// @expect: ababab ""  xx ""
// @expect: RangeError Invalid count value: -1
// @expect: RangeError Invalid count value: Infinity
// @expect: RangeError Invalid count value: -Infinity
// @expect: RangeError Invalid string length
console.log("ab".repeat(3), JSON.stringify("".repeat(2147483647)), "x".repeat(0), "x".repeat(2.9), JSON.stringify("x".repeat(-0.5)));
for (const c of [-1, Infinity, -Infinity]) {
  try { "a".repeat(c); } catch (e) { console.log((e as Error).name, (e as Error).message); }
}
try { "abc".repeat(2 ** 30); } catch (e) { console.log((e as Error).name, (e as Error).message); }
