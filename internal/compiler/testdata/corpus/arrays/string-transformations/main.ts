const words: string[] = ["apple", "banana", "cherry", "date"];

console.log(words.length);
console.log(words.join(" | "));

words.push("elderberry");
console.log(words.join(","));

const removedLast: string = words.pop()!;
console.log(removedLast);

const removedFirst: string = words.shift()!;
console.log(removedFirst);

words.unshift("avocado");
console.log(words.join(","));

const longWords: string[] = words.filter((w: string): boolean => w.length > 5);
console.log(longWords.join(";"));

const upperWords: string[] = words.map((w: string): string => w.toUpperCase());
console.log(upperWords.join(" "));

console.log(words.includes("banana"));
console.log(words.includes("mango"));
console.log(words.indexOf("cherry"));
console.log(words.indexOf("nonexistent"));
