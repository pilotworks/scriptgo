const nums: number[] = [10, 20, 30];
console.log(nums.push(40));
console.log(nums.length);
console.log(nums.pop());
console.log(nums.length);
console.log(nums.indexOf(20));
console.log(nums.indexOf(99));
console.log(nums.includes(30));
console.log(nums.includes(99));

const sub: number[] = nums.slice(1, 3);
console.log(sub.length);
console.log(sub[0]);
console.log(sub[1]);
