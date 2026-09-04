// Dynamic object keys must use the runtime property name for both writes and reads.
const values: Record<string, string> = {};
const first = "first";
const second = "second";
values[first] = "one";
values[second] = "two";
console.log(values[first]);
console.log(values[second]);
console.log(Object.keys(values).length);

// @run.expected: one
// @run.expected: two
// @run.expected: 2
