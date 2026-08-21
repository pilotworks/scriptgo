const encoder = new TextEncoder();
const decoder = new TextDecoder();
const u8 = encoder.encode("hello decoder");
console.log(decoder.decode(u8));
