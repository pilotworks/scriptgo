const d1 = new Date(1700000000000);
console.log(d1.getTime());
console.log(d1.toISOString());

const d2 = new Date("2023-11-14T22:13:20.000Z");
console.log(d2.getTime());
console.log(d2.toISOString());

const parsed: number = Date.parse("2023-11-14T22:13:20.000Z");
console.log(parsed);

const now: number = Date.now();
console.log(now > 1600000000000);
