const text = "ScriptGo Compiler Engine";

console.log(text.startsWith("Script"));
console.log(text.startsWith("Engine"));
console.log(text.endsWith("Engine"));
console.log(text.endsWith("Script"));

console.log(text.includes("Compiler"));
console.log(text.includes("xyz"));

console.log(text.indexOf("Go"));
console.log(text.indexOf("xyz"));
console.log(text.lastIndexOf("e"));

console.log(text.charAt(0));
console.log(text.charCodeAt(0));

console.log("abc".repeat(3));
console.log("  padded  ".trimStart());
console.log("  padded  ".trimEnd());
console.log(text.toLowerCase());
console.log(text.toUpperCase());
