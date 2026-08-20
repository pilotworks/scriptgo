const d = new Date(1700000000000);
console.log(d.getTime());
console.log(d.toISOString());
console.log(d.toString().length > 0);

const ts = Date.parse("2024-01-01T00:00:00.000Z");
console.log(ts);

const d2 = new Date(ts);
console.log(d2.toISOString());
