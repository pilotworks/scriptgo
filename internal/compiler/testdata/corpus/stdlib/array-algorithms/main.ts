const arr: number[] = [10, 20, 30, 40, 50];

console.log(arr.indexOf(30));
console.log(arr.indexOf(99));
console.log(arr.includes(40));
console.log(arr.includes(100));

const joined = arr.join("-");
console.log(joined);

const mapped = arr.map((x: number): number => x * 2);
console.log(mapped.join(","));

const filtered = arr.filter((x: number): boolean => x > 25);
console.log(filtered.join(","));

const total = arr.reduce((acc: number, curr: number): number => acc + curr, 0);
console.log(total);

const found = arr.find((x: number): boolean => x > 25);
console.log(found);

const hasEven = arr.some((x: number): boolean => x % 2 === 0);
console.log(hasEven);

const allPositive = arr.every((x: number): boolean => x > 0);
console.log(allPositive);
