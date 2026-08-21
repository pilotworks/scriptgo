console.log(Number.isInteger(42));
console.log(Number.isInteger(42.5));
console.log(Number.isInteger(NaN));
console.log(Number.isInteger(Infinity));

console.log(Number.isFinite(100));
console.log(Number.isFinite(Infinity));
console.log(Number.isFinite(NaN));

console.log(Number.isNaN(NaN));
console.log(Number.isNaN(123));

console.log(Number.parseInt("12345"));
console.log(Number.parseFloat("3.14159"));

console.log(Number.MAX_SAFE_INTEGER);
console.log(Number.MIN_SAFE_INTEGER);
console.log(Number.EPSILON > 0 && Number.EPSILON < 1e-15);
