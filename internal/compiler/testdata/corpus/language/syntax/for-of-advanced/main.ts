// 1. Iterate over string characters
const text: string = "abc";
for (const ch of text) {
    console.log(ch);
}

// 2. Iterate over boolean array
const flags: boolean[] = [true, false, true];
for (const f of flags) {
    console.log(f);
}

// 3. Iterate over object array
class Item {
    id: number;
    name: string;
    constructor(id: number, name: string) {
        this.id = id;
        this.name = name;
    }
}

const items: Item[] = [new Item(1, "A"), new Item(2, "B")];
for (const item of items) {
    console.log(item.id);
    console.log(item.name);
}

// 4. Continue and break in for...of
const nums: number[] = [10, 20, 30, 40, 50];
let sum: number = 0;
for (const n of nums) {
    if (n === 20) {
        continue;
    }
    if (n === 50) {
        break;
    }
    sum += n;
}
console.log(sum); // 10 + 30 + 40 = 80
