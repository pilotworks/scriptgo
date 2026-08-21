const id: string = "42";
console.log(id.padStart(5, "0"));
console.log(id.padEnd(6, "-"));

const raw: string = "   leading and trailing   ";
console.log(`[${raw.trimStart()}]`);
console.log(`[${raw.trimEnd()}]`);

const rep: string = "abc-";
console.log(rep.repeat(3));

const greeting: string = "Hello";
console.log(greeting.charAt(0));
console.log(greeting.charAt(4));
console.log(greeting.charCodeAt(0));
console.log(greeting.charCodeAt(1));
