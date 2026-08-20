const u1: unknown = 100;
const u2: unknown = "scriptgo";
const u3: unknown = false;

const n: number = u1 as number;
const s: string = u2 as string;
const b: boolean = u3 as boolean;

console.log(n + 50);
console.log(s);
console.log(b);
