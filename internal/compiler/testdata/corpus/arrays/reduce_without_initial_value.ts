// @expect: 10
// @expect: 4321
// @expect: 110
// @expect: threw
// reduce/reduceRight without an initial value seed from the first visited
// element and throw a TypeError on an empty array.
const values: number[] = [1, 2, 3, 4];
console.log(values.reduce((acc, value) => acc + value));
console.log(values.reduceRight((acc, value) => acc * 10 + value));
console.log(values.reduce((acc, value) => acc + value, 100));
const empty: number[] = [];
try {
  empty.reduce((acc, value) => acc + value);
  console.log("not thrown");
} catch (error) {
  console.log("threw");
}
