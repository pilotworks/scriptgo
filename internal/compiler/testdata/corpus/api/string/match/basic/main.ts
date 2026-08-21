const s: string = "hello 123";
const m = s.match(/[0-9]+/) as string[];
console.log(m[0]);
