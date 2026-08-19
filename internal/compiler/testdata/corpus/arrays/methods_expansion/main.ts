const nums: number[] = [10, 20, 30, 40];
console.log(nums.at(0)!);
console.log(nums.at(-1)!);
console.log(nums.join("-"));

const first: number = nums.shift()!;
console.log(first);
console.log(nums.length);

const newLen: number = nums.unshift(5);
console.log(newLen);
console.log(nums.at(0)!);

const rev: number[] = nums.reverse();
console.log(rev.join(","));

const more: number[] = [100, 200];
const combined: number[] = nums.concat(more);
console.log(combined.join(" "));

const strs: string[] = ["a", "b", "c", "d"];
console.log(strs.at(-2)!);
const spliced: string[] = strs.splice(1, 2);
console.log(spliced.join(":"));
console.log(strs.join(":"));
