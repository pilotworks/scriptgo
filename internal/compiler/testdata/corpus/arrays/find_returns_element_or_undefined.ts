// @expect: 15 21 undefined
// @expect: banana apple undefined
// @expect: true
// @expect: 9
// @expect: missing
// find/findLast return the element or undefined for every element type.
const nums: number[] = [3, 8, 15, 21];
console.log(nums.find((n) => n > 10), nums.findLast((n) => n > 10), nums.find((n) => n > 100));
const words: string[] = ["apple", "banana", "cherry"];
console.log(words.find((w) => w.startsWith("b")), words.findLast((w) => w.length === 5), words.find((w) => w === "kiwi"));
const flags: boolean[] = [false, true];
console.log(flags.find((f) => f));
const first: number = nums.find((n) => n % 2 === 0)!;
console.log(first + 1);
const maybe = nums.find((n) => n === 99);
console.log(maybe === undefined ? "missing" : "present");
