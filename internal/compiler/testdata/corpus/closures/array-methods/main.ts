const nums = [1, 2, 3, 4, 5];
const factor = 2;
const doubled = nums.map((x) => x * factor);
console.log(doubled.join(", "));

const evens = nums.filter((x) => x % 2 === 0);
console.log(evens.join(", "));

nums.forEach((x) => {
    if (x > 3) {
        console.log(x);
    }
});

const sum = nums.reduce((acc, x) => acc + x, 0);
console.log(sum);

const found = nums.find((x) => x === 3);
console.log(found);

console.log(nums.some((x) => x > 4));
console.log(nums.every((x) => x > 0));
