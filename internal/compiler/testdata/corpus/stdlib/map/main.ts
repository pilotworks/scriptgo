const m = new Map<string, number>();
m.set("apple", 10);
m.set("banana", 20);
m.set("orange", 30);

console.log(m.size);
console.log(m.get("banana"));
console.log(m.has("apple"));
console.log(m.has("grape"));

m.delete("banana");
console.log(m.size);
console.log(m.has("banana"));

m.forEach((val, key) => {
    console.log(key, val);
});

m.clear();
console.log(m.size);

const numMap = new Map<number, string>();
numMap.set(100, "hundred");
numMap.set(200, "two hundred");
console.log(numMap.get(100));
console.log(numMap.get(200));
console.log(numMap.has(100));
console.log(numMap.has(300));
console.log(numMap.size);
