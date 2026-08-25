// @expect: apple: 5
// @expect: banana: 10
// @expect: cherry: 15
const pairs: [string, number][] = [
    ["apple", 5],
    ["banana", 10],
    ["cherry", 15]
];

for (const [name, count] of pairs) {
    console.log(name + ": " + count);
}
