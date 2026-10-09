// @expect: plain true
// @expect: gen true
// @expect: 0
// A parameter whose type is void or never (its default always throws) is
// stored boxed; calling without it throws before the body runs.
let calls = 0;
function fail(): never {
  throw new RangeError("no default");
}
function plain(_ = fail()): void {
  calls = calls + 1;
}
function* gen(_ = fail()) {
  calls = calls + 1;
  yield 1;
}
try { plain(); } catch (e) { console.log("plain", e instanceof RangeError); }
try { gen(); } catch (e) { console.log("gen", e instanceof RangeError); }
console.log(calls);
