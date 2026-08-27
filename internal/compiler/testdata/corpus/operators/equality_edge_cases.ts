// @expect: true
// @expect: false
// @expect: true
// @expect: false
// @expect: true
// @expect: true
// @expect: false
const a: number = 0;
const b: number = 0;
console.log(a === b);
console.log(a !== b);

const str1: string = "hello";
const str2: string = "hello";
console.log(str1 === str2);

const str3: string = "world";
console.log(str1 === str3);

const b1: boolean = true;
const b2: boolean = true;
console.log(b1 === b2);

const n1: number | null = null;
const n2: number | null = null;
console.log(n1 === n2);

const numVal: number | null = 42;
console.log(numVal === n1);
