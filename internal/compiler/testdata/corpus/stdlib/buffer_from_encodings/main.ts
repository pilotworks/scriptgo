const bUtf8: Buffer = Buffer.from("hello", "utf8");
console.log(bUtf8.length);
console.log(bUtf8.toString());

const bHex: Buffer = Buffer.from("68656c6c6f", "hex");
console.log(bHex.toString());
console.log(bHex.toString("hex"));

const bB64: Buffer = Buffer.from("aGVsbG8=", "base64");
console.log(bB64.toString());
console.log(bB64.toString("base64"));

console.log(Buffer.byteLength("hello"));
console.log(Buffer.byteLength("68656c6c6f", "hex"));
console.log(Buffer.byteLength("aGVsbG8=", "base64"));
