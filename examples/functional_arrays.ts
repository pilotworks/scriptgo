// Functional Array methods demo: map, filter, reduce, find, some, every, forEach

console.log("=== Functional Array Methods Demo ===");

const nums: readonly number[] = [1, 2, 3, 4, 5, 6];
const factor: number = 2;

// 1. map
const doubled: number[] = nums.map((x: number): number => x * factor);
console.log(`Doubled: ${doubled.join(", ")}`);

// 2. filter
const evens: number[] = nums.filter((x: number): boolean => x % 2 === 0);
console.log(`Evens: ${evens.join(", ")}`);

// 3. reduce
const sum: number = nums.reduce((acc: number, x: number): number => acc + x, 0);
console.log(`Sum: ${sum}`);

// 4. find
const found: number | undefined = nums.find((x: number): boolean => x === 4);
console.log(`Found 4: ${found}`);

// 5. some & every
console.log(`Has any > 5: ${nums.some((x: number): boolean => x > 5)}`);
console.log(`All positive: ${nums.every((x: number): boolean => x > 0)}`);
