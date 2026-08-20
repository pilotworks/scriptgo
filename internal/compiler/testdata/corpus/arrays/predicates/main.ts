const nums: number[] = [10, 15, 20, 25, 30];

const hasOdd: boolean = nums.some((n: number): boolean => n % 2 !== 0);
console.log(hasOdd);

const hasNegative: boolean = nums.some((n: number): boolean => n < 0);
console.log(hasNegative);

const allPositive: boolean = nums.every((n: number): boolean => n > 0);
console.log(allPositive);

const allEven: boolean = nums.every((n: number): boolean => n % 2 === 0);
console.log(allEven);

const foundFirstOverTwenty: number = nums.find((n: number): boolean => n > 20)!;
console.log(foundFirstOverTwenty);

const foundFirstDivBySix: number = nums.find((n: number): boolean => n % 6 === 0)!;
console.log(foundFirstDivBySix);
