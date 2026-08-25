// @expect: 1,3,5
// @expect: two,four
type NumItem = { kind: "num"; val: number };
type StrItem = { kind: "str"; val: string };
type Item = NumItem | StrItem;

const mixed: Item[] = [
    { kind: "num", val: 1 },
    { kind: "str", val: "two" },
    { kind: "num", val: 3 },
    { kind: "str", val: "four" },
    { kind: "num", val: 5 }
];

const nums: number[] = [];
const strs: string[] = [];

for (const item of mixed) {
    if (item.kind === "num") {
        nums.push((item as NumItem).val);
    } else {
        strs.push((item as StrItem).val);
    }
}

console.log(nums.join(","));
console.log(strs.join(","));
