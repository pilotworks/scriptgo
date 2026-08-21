const wrap = <T>(val: T): T[] => [val];

const nums = wrap(100);
console.log(nums[0]);

const strs = wrap("hello");
console.log(strs[0]);
