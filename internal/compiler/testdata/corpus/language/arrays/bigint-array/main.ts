const bigs: bigint[] = [10n, 20n, 30n];
console.log(bigs.length);
console.log(bigs[0]);
console.log(bigs[1]);
console.log(bigs[2]);

bigs[1] = 99n;
console.log(bigs[1]);

console.log(bigs.push(40n));
console.log(bigs.length);
console.log(bigs.pop()!);
console.log(bigs.length);
console.log(bigs.join(","));
