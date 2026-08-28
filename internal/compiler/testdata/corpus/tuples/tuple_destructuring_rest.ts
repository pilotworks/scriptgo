// @expect: scriptgo
// @expect: 2026
// @expect: true
// 1. Tuple destructuring with rest elements and type preservation
type MixedTuple = [string, number, boolean];
const sampleTuple: MixedTuple = ["scriptgo", 2026, true];

const [firstItem, ...restItems] = sampleTuple;
console.log(firstItem);
console.log(restItems[0]);
console.log(restItems[1]);
