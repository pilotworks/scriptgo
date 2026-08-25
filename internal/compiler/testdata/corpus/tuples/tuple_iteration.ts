// @expect: first
// @expect: second
// @expect: third
const tup: [string, string, string] = ["first", "second", "third"];

for (const el of tup) {
    console.log(el);
}
