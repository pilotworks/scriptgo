// @expect: 14
// @expect: 84
const unknownVal: unknown = "hello scriptgo";
const strLen = (unknownVal as string).length;
console.log(strLen);

const numVal: unknown = 42;
const doubled = (numVal as number) * 2;
console.log(doubled);
