const a: number[] = [1, 2, 3, 4];
const sum = a.reduceRight((acc: number, val: number) => acc + val, 0);
console.log(sum);
