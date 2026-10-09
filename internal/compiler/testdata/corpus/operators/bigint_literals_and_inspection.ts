// @expect: 16n 1000n 5n 15n
// @expect: [ 16n, -3n ]
// @expect: value=16n
// @expect: 16 16 16
// BigInt literals in every radix normalize to 64-bit values, and console
// formatting inspects bigints with an "n" suffix.
const hex = 0x10n;
const separated = 1_000n;
const binary = 0b101n;
const octal = 0o17n;
console.log(hex, separated, binary, octal);
console.log([hex, -3n]);
console.log("value=%s", hex);
console.log(`${hex}`, String(hex), hex.toString());
