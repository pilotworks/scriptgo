let sum = 0;
const promises = [Promise.resolve(10), Promise.resolve(20), Promise.resolve(30)];
for await (const x of promises) {
    sum += x;
    console.log(x);
}
console.log(sum);

let numSum = 0;
const nums = [1, 2, 3, 4];
for await (const n of nums) {
    numSum += n;
}
console.log(numSum);
export {};
