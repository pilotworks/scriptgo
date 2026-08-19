type NumArray = Array<number>;
type StrArray = Array<string>;
type MyList<T> = T[];

const a: NumArray = [1, 2, 3];
const b: StrArray = ["hello", "world"];
const c: MyList<number> = [10, 20, 30];

console.log(a.length);
console.log(b[0]);
console.log(c[2]);
