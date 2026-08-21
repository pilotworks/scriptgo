const a: number[] = [1, 2, 3];
const res = a.flatMap((x: number) => x * 2);
console.log(res.join(","));
