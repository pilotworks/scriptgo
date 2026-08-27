// @expect: 14
// @expect: 20
// @expect: 16
// @expect: 42
// @expect: default_val
// @expect: 0
const a = 2 + 3 * 4;
const b = (2 + 3) * 4;
console.log(a);
console.log(b);

const c = 1 << (2 + 2); // 1 << 4 = 16
console.log(c);

const x: number | null = null;
const y: number | null = 42;
const z = x ?? y ?? 100;
console.log(z);

const str: string | null = null;
console.log(str ?? "default_val");

const zero: number = 0;
console.log(zero ?? 999);
