const oneByte = Buffer.from("TQ==", "base64");
const twoBytes = Buffer.from("TWE=", "base64");
const threeBytes = Buffer.from("TWFu", "base64");
const vectorized = Buffer.from("QUJDREVGR0hJSktMTU5PUFFSU1RVVldYWVoxMjM0NTY3ODkwYWJjZGVm", "base64");

console.log(oneByte.length, oneByte.toString());
console.log(twoBytes.length, twoBytes.toString());
console.log(threeBytes.length, threeBytes.toString());
console.log(vectorized.length, vectorized.toString());
