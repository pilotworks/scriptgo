// Functional Array methods demo: map, filter, reduce, find, some, every, forEach

console.log("=== Functional Array Methods Demo ===");

const nums = [1, 2, 3, 4, 5, 6];
const factor = 2;

// 1. map
const doubled = nums.map((x) => x * factor);
console.log("Doubled: " + doubled.join(", "));

// 2. filter
const evens = nums.filter((x) => x % 2 === 0);
console.log("Evens: " + evens.join(", "));

// 3. reduce
const sum = nums.reduce((acc, x) => acc + x, 0);
console.log("Sum: " + sum);

// 4. find
const found = nums.find((x) => x === 4);
console.log("Found 4: " + found);

// 5. some & every
console.log("Has any > 5: " + nums.some((x) => x > 5));
console.log("All positive: " + nums.every((x) => x > 0));
