const s = new Set<number>();
s.add(10);
s.add(20);
s.add(30);
s.add(20);

console.log(s.size);
console.log(s.has(20));
console.log(s.has(40));

s.delete(20);
console.log(s.size);
console.log(s.has(20));

s.forEach((val) => {
    console.log(val);
});

s.clear();
console.log(s.size);

const strSet = new Set<string>();
strSet.add("foo");
strSet.add("bar");
strSet.add("foo");
console.log(strSet.size);
console.log(strSet.has("foo"));
console.log(strSet.has("baz"));
