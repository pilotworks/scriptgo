const nums: number[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];

const evens: number[] = nums.filter((n: number): boolean => n % 2 === 0);
console.log(evens.join(","));

const doubled: number[] = evens.map((n: number): number => n * 2);
console.log(doubled.join(","));

const sliced: number[] = doubled.slice(1, 4);
console.log(sliced.join(","));

const more: number[] = [100, 200];
const combined: number[] = sliced.concat(more);
console.log(combined.join(","));

const reversed: number[] = combined.reverse();
console.log(reversed.join(","));
