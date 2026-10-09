// @expect: abc cba
// @expect: 6 123
// @expect: n1,n2,n3
// @expect: 8 3
// @expect: true
// @expect: init
// reduce/reduceRight with string, array, and mixed accumulators, with and
// without an initial value (an empty array without one throws TypeError).
const words = ["a", "b", "c"];
console.log(words.reduce((x, y) => x + y), words.reduceRight((x, y) => x + y));
const nums = [1, 2, 3];
console.log(nums.reduce((acc, v) => acc + v, 0), nums.reduce((acc, v) => acc + String(v), ""));
const collected: string[] = nums.reduce((acc: string[], v) => {
  acc.push("n" + v);
  return acc;
}, []);
console.log(collected.join());
const total = nums.reduceRight((acc, v, i) => acc + v * i, 0);
console.log(total, nums.reduce((a, b) => Math.max(a, b)));
const empty: string[] = [];
try {
  empty.reduce((x, y) => x + y);
} catch (e) {
  console.log(e instanceof TypeError);
}
console.log(empty.reduce((x, y) => x + y, "init"));
