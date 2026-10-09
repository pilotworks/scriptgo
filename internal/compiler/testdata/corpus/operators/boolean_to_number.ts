// @ts-nocheck
// @expect: 1 2 2 0 xtrue truey
// @expect: true true true 3 8 -1 0
// @expect: false false Infinity 1
// @expect: -2 -0 1
// Booleans in arithmetic, bitwise and relational operators use ToNumber
// (true is 1, false is 0). TypeScript rejects these (TS2365), so this file
// opts out of type checking the way test262 conformance code does.
const a = true; const b = false;
console.log(a + b, true + 1, 2 * true, a - 1, "x" + a, a + "y");
console.log(true > false, true >= 1, false < 0.5, true | 2, true << 3, -true, +false);
console.log(a === b, a == b, true / false, false ** 0);
console.log(~true, -false, +true);
