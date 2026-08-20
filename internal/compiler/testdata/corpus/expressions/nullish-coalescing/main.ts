const a: string = "hello";
const b: string = a ?? "fallback";
console.log(b);

const undef: string | undefined = undefined;
const c: string = undef ?? "fallback";
console.log(c);

const nul: string | null = null;
const d: string = nul ?? "fallback";
console.log(d);

let num: number = 256;
num >>>= 3;
console.log(num);
