// @expect: 1
// `[] as T[]` is an empty T[] that methods such as push can grow.
const b = { parts: [] as string[] };
b.parts.push("x");
console.log(b.parts.length);
